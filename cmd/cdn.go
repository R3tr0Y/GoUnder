package cmd

import (
	"GoUnder/utils"
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

func (f *FofaResponse) UnmarshalJSON(data []byte) error {
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

	var result2D [][]string
	if err := json.Unmarshal(aux.Results, &result2D); err == nil {
		f.Results = result2D
		return nil
	}

	var result1D []string
	if err := json.Unmarshal(aux.Results, &result1D); err == nil {
		for _, r := range result1D {
			f.Results = append(f.Results, []string{r})
		}
		return nil
	}

	return fmt.Errorf("无法解析 results 字段: %s", string(aux.Results))
}

var targetURL string
var pattern string
var cfg *FofaConfig

var cdnCmd = &cobra.Command{
	Use:   "cdn",
	Short: "Seek true IP behind CDN servers.",
	Run: func(cmd *cobra.Command, args []string) {
		if targetURL == "" {
			fmt.Println("❗  use -u for target URL")
			_ = cmd.Usage()
			os.Exit(1)
		}
		cdnLookup(targetURL)
	},
}

func cdnLookup(input string) {
	var err error
	cfg, err = loadFofaConfig()
	if err != nil {
		fmt.Println("Read config file error:", err)
		os.Exit(1)
	}
	fmt.Println("Fofa account loaded:", cfg.Email)

	patterns := []string{"host", "title", "icon"}
	if pattern != "" {
		patterns = []string{pattern}
	}

	resultSet := make(map[string]bool)

	for _, p := range patterns {
		queries, encoded := get_queries(p, input)
		if queries != nil {
			fmt.Println("Query string:", queries)
		}
		for _, enc := range encoded {
			for _, ip := range Query(enc, "ip,port,host") {
				resultSet[ip] = true
			}
		}
	}

	fmt.Println("\n✅ Promising true IP & Ports found: ")
	for ip := range resultSet {
		fmt.Println("-", ip)
	}
}

func get_queries(p string, input string) ([]string, []string) {
	var queries, encodedQueries []string

	switch p {
	case "host":
		q := fmt.Sprintf(`host="%s" && is_cloud=false`, extractHost(input))
		queries = append(queries, q)

	case "title":
		titles, _ := get_titles(input)
		for _, title := range titles {
			fmt.Println("Get website title:", title)
			q := fmt.Sprintf(`title="%s" && is_cloud=false`, title)
			queries = append(queries, q)
		}

	case "icon":
		iconHash, err := getFaviconHash(input)
		if err != nil {
			fmt.Println("获取 icon_hash 失败:", err)
			break
		}
		fmt.Println("Favicon hash:", iconHash)
		q := fmt.Sprintf(`icon_hash="%s" && is_cloud=false`, iconHash)
		queries = append(queries, q)
	}

	for _, q := range queries {
		encodedQueries = append(encodedQueries, base64.StdEncoding.EncodeToString([]byte(q)))
	}
	return queries, encodedQueries
}

func get_titles(url string) ([]string, error) {
	var titles []string
	seen := make(map[string]bool)

	// 构造 FOFA 查询
	query := fmt.Sprintf(`host="%s"`, extractHost(url))
	encodedQuery := base64.StdEncoding.EncodeToString([]byte(query))

	// 调用 FOFA 查询 title 字段
	results := Query(encodedQuery, "title")
	for _, title := range results {
		trimmed := strings.TrimSpace(title)
		if trimmed != "" && !seen[trimmed] {
			titles = append(titles, trimmed)
			seen[trimmed] = true
		}
	}

	// 本地抓取网页 title 标签
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	client := resty.New()
	resp, err := client.R().Get(url)
	if err != nil && strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "https://", 1)
		resp, err = client.R().Get(url)
	}
	if err == nil {
		body := resp.String()
		start := strings.Index(body, "<title>")
		end := strings.Index(body, "</title>")
		if start != -1 && end != -1 && start < end {
			title := strings.TrimSpace(body[start+len("<title>") : end])
			if title != "" && !seen[title] {
				titles = append(titles, title)
			}
		}
	}

	if len(titles) == 0 {
		return nil, fmt.Errorf("未能获取任何有效的 title")
	}

	return titles, nil
}
func getFaviconHash(input string) (string, error) {
	host := extractHost(input)
	url := host
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	// 下载 favicon
	favURL := url + "/favicon.ico"
	// client := resty.New()
	// resp, err := client.R().Get(favURL)
	// if err != nil {
	// 	// fallback to https
	// 	favURL = strings.Replace(favURL, "http://", "https://", 1)
	// 	resp, err = client.R().Get(favURL)
	// 	if err != nil {
	// 		return "", fmt.Errorf("无法获取 favicon.ico: %v", err)
	// 	}
	// }

	hash, _ := utils.GetIconHashFromURL(favURL)
	fmt.Println(hash)
	return fmt.Sprintf("%v", hash), nil // FOFA 使用的是有符号 int32
}
func extractHost(raw string) string {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return raw
	}
	u, _ := url.Parse(raw)
	return u.Host
}

func loadFofaConfig() (*FofaConfig, error) {
	path := filepath.Join("configs", "fofa.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
}

func Query(encodedQuery string, fields ...string) []string {
	client := resty.New()
	var result FofaResponse
	f := ""
	if len(fields) > 0 {
		f = fields[0]
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

	results := make(map[string]bool)
	for _, entry := range result.Results {
		if len(entry) > 0 && entry[0] != "" {
			results[strings.Join(entry, ", ")] = true
		}
	}

	var unique []string
	for ip := range results {
		unique = append(unique, ip)
	}
	return unique
}

func init() {
	cdnCmd.Flags().StringVarP(&targetURL, "url", "u", "", "目标URL（如：https://example.com）")
	cdnCmd.Flags().StringVarP(&pattern, "pattern", "p", "", "[host | title | icon], default: all")
	rootCmd.AddCommand(cdnCmd)
}
