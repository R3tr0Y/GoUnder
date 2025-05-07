package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

type FofaConfig struct {
	Email string `json:"email"`
	Key   string `json:"key"`
}

type FofaResponse struct {
	Error   bool       `json:"error"`
	Results [][]string `json:"results"`
	Msg     string     `json:"errmsg"`
}

var targetURL string
var pattern string
var cfg FofaConfig
var cdnCmd = &cobra.Command{
	Use:   "cdn",
	Short: "尝试识别隐藏在CDN后的真实IP",
	Run: func(cmd *cobra.Command, args []string) {
		if targetURL == "" {
			fmt.Println("❗ 请使用 -u 指定目标 URL")
			cmd.Usage()
			os.Exit(1)
		}

		// if !isValidURL(targetURL) {
		// 	fmt.Println("❗ 提供的 URL 无效:", targetURL)
		// 	os.Exit(1)
		// }

		cfg, err := loadFofaConfig()
		if err != nil {
			fmt.Println("配置文件读取失败:", err)
			os.Exit(1)
		} else {
			fmt.Println("fofa account loaded: " + cfg.Email)
		}

		// query := fmt.Sprintf(`host="%s" && is_cloud=false`, extractHost(targetURL))
		// encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
		var query, encodedQuery string
		if pattern == "" {
			query, encodedQuery = get_query("host", targetURL)
		} else {
			query, encodedQuery = get_query(pattern, targetURL)
		}

		fmt.Println("Query string: " + query)
		fmt.Println("Base64 encoded: " + encodedQuery)

		// os.Exit(1)
		Query(encodedQuery)
	},
}

func isValidURL(raw string) bool {
	u, err := url.ParseRequestURI(raw)
	return err == nil && u.Host != ""
}

func extractHost(raw string) string {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return raw
	} else {
		u, _ := url.Parse(raw)
		return u.Host
	}
}

func loadFofaConfig() (*FofaConfig, error) {
	path := filepath.Join("configs", "fofa.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// var cfg FofaConfig
	err = json.Unmarshal(data, &cfg)
	return &cfg, err
}

func get_query(p string, input string) (string, string) {
	switch p {
	case "host":
		query := fmt.Sprintf(`host="%s" && is_cloud=false`, extractHost(input))
		encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
		return query, encodedQuery
	case "title":
		title, _ := get_title(input)
		fmt.Println("Get website title: " + title)
		query := fmt.Sprintf(`title="%s" && is_cloud=false`, title)
		encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
		return query, encodedQuery
	case "icon":
		query := fmt.Sprintf(`icon_hash="%s" && is_cloud=false`, extractHost(input))
		encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
		return query, encodedQuery
	}
	return "", ""
}

func get_title(url string) (string, error) {
	// 确保URL有协议前缀
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url // 默认使用http，如果失败可以尝试https
	}

	client := resty.New()
	resp, err := client.R().Get(url)
	if err != nil {
		// 尝试https
		if strings.HasPrefix(url, "http://") {
			url = strings.Replace(url, "http://", "https://", 1)
			resp, err = client.R().Get(url)
			if err != nil {
				return "", fmt.Errorf("failed to fetch URL: %v", err)
			}
		} else {
			return "", fmt.Errorf("failed to fetch URL: %v", err)
		}
	}

	body := resp.String()
	titleStart := strings.Index(body, "<title>")
	if titleStart == -1 {
		return "", fmt.Errorf("no title found")
	}
	titleStart += len("<title>")
	titleEnd := strings.Index(body, "</title>")
	if titleEnd == -1 {
		return "", fmt.Errorf("malformed title tag")
	}
	query := fmt.Sprintf(`host="%s"`, url)
	encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
	Query(encodedQuery)
	return body[titleStart:titleEnd], nil

}

func Query(encodedQuery string, field ...string) map[string]bool {
	client := resty.New()
	var result FofaResponse
	f := ""
	if len(field) > 0 {
		f = field[0]
	}
	_, err := client.R().
		SetQueryParams(map[string]string{
			"email":   cfg.Email,
			"key":     cfg.Key,
			"qbase64": encodedQuery,
			"size":    "100",
			"field":   f,
		}).
		SetResult(&result).
		Get("https://fofa.info/api/v1/search/all")

	if err != nil {
		fmt.Println("请求 FOFA API 失败:", err)
		return nil
	}

	if result.Error {
		fmt.Printf("FOFA 返回错误: %s\n", result.Msg)
		return nil
	}

	ips := make(map[string]bool)
	for _, entry := range result.Results {
		if len(entry) > 0 {
			ipPort := entry[0]
			ips[ipPort] = true
		}
	}

	if len(ips) == 0 {
		fmt.Println("未发现任何非CDN的IP")
		return nil
	}

	fmt.Println("✅ 找到以下可能的真实IP及开放的端口：")
	for ip := range ips {
		fmt.Println(" -", ip)
	}
	return ips
}

func init() {
	cdnCmd.Flags().StringVarP(&targetURL, "url", "u", "", "目标URL（如：https://example.com）")
	cdnCmd.Flags().StringVarP(&pattern, "pattern", "p", "", "[host | title | icon], default: all")
	rootCmd.AddCommand(cdnCmd)
}
