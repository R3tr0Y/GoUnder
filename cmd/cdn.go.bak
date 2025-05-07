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
	Error   bool            `json:"error"`
	Results [][]string      `json:"-"`
	Msg     string          `json:"errmsg"`
	raw     json.RawMessage `json:"results"`
}

// 自定义 UnmarshalJSON 方法处理兼容结构
func (f *FofaResponse) UnmarshalJSON(data []byte) error {
	// 定义辅助结构体匹配大结构
	type Alias FofaResponse
	aux := &struct {
		Results json.RawMessage `json:"results"`
		*Alias
	}{
		Alias: (*Alias)(f),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 尝试先当成 [][]string 来解析
	var result2D [][]string
	if err := json.Unmarshal(aux.Results, &result2D); err == nil {
		f.Results = result2D
		return nil
	}

	// 如果不是 [][]string，再试着按 []string 解析
	var result1D []string
	if err := json.Unmarshal(aux.Results, &result1D); err == nil {
		// 包装成 [][]string 的形式
		for _, r := range result1D {
			f.Results = append(f.Results, []string{r})
		}
		return nil
	}

	// 两种都不匹配，返回错误
	return fmt.Errorf("无法解析 results 字段: %s", string(aux.Results))
}

var targetURL string
var pattern string
var cfg FofaConfig
var real_ips []string

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
		var queries, encodedQueries []string
		if pattern != "" {
			queries, encodedQueries = get_queries(pattern, targetURL)
			fmt.Println("Query string: ", queries)
			for _, encodedQuery := range encodedQueries {
				real_ips = append(Query(encodedQuery), real_ips...)
			}
		} else {
			queries, encodedQueries = get_queries("host", targetURL)
			fmt.Println("Query string: ", queries)
			for _, encodedQuery := range encodedQueries {
				real_ips = append(Query(encodedQuery), real_ips...)
			}
			queries, encodedQueries = get_queries("title", targetURL)
			fmt.Println("Query string: ", queries)
			for _, encodedQuery := range encodedQueries {
				real_ips = append(Query(encodedQuery), real_ips...)
			}
			queries, encodedQueries = get_queries("icon", targetURL)
			fmt.Println("Query string: ", queries)
			for _, encodedQuery := range encodedQueries {
				real_ips = append(Query(encodedQuery), real_ips...)
			}

		}

		// fmt.Println("Query string: ", queries)
		// fmt.Println("Base64 encoded: ", encodedQueries)

		// os.Exit(1)

		fmt.Println("✅ 找到以下可能的真实IP及开放的端口：")
		for _, ip := range real_ips {
			fmt.Println("- " + ip)

		}
	},
}

// func isValidURL(raw string) bool {
// 	u, err := url.ParseRequestURI(raw)
// 	return err == nil && u.Host != ""
// }

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

func get_queries(p string, input string) ([]string, []string) {
	var queries, encodedQueries []string
	switch p {
	case "host":
		query := fmt.Sprintf(`host="%s" && is_cloud=false`, extractHost(input))
		encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
		queries = append(queries, query)
		encodedQueries = append(encodedQueries, encodedQuery)
		return queries, encodedQueries
	case "title":
		titles, _ := get_titles(input)
		if len(titles) > 0 {
			for _, title := range titles {
				fmt.Println("Get website title: " + title)
				query := fmt.Sprintf(`title="%s" && is_cloud=false`, title)
				encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
				queries = append(queries, query)
				encodedQueries = append(encodedQueries, encodedQuery)
			}
			return queries, encodedQueries
		} else {
			return nil, nil
		}

		// title, _ := get_titles(input)

	case "icon":
		query := fmt.Sprintf(`icon_hash="%s" && is_cloud=false`, extractHost(input))
		encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))
		queries = append(queries, query)
		encodedQueries = append(encodedQueries, encodedQuery)
		return queries, encodedQueries
	}
	return nil, nil
}

func get_titles(url string) ([]string, error) {
	var titles []string
	query := fmt.Sprintf(`host="%s"`, url)
	encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))

	q := Query(encodedQuery, "title")

	if q != nil {
		titles = append(titles, q...)
	}

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
				return titles, fmt.Errorf("failed to fetch URL: %v", err)
			}
		} else {
			return titles, fmt.Errorf("failed to fetch URL: %v", err)
		}
	}

	body := resp.String()
	titleStart := strings.Index(body, "<title>")
	if titleStart == -1 {
		return titles, fmt.Errorf("no html title found")
	}
	titleStart += len("<title>")
	titleEnd := strings.Index(body, "</title>")
	if titleEnd == -1 {
		return titles, fmt.Errorf("malformed html title tag")
	}
	titles = append(titles, body[titleStart:titleEnd])

	return titles, nil

}

func Query(encodedQuery string, field ...string) []string {
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
			"fields":  f,
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

	results := make(map[interface{}]bool)
	for _, entry := range result.Results {
		if len(entry) > 0 {
			ipPort := entry[0]
			results[ipPort] = true

		}
	}

	if len(results) == 0 {
		// fmt.Println("Not found any results")
		return nil
	}
	var ips_slice []string
	for r := range results {
		// fmt.Println(" -", ip)
		if r != "" {
			ips_slice = append(ips_slice, r.(string))
		}

	}
	return ips_slice
}

func init() {
	cdnCmd.Flags().StringVarP(&targetURL, "url", "u", "", "目标URL（如：https://example.com）")
	cdnCmd.Flags().StringVarP(&pattern, "pattern", "p", "", "[host | title | icon], default: all")
	rootCmd.AddCommand(cdnCmd)
}
