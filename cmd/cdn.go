package cmd

import (
	"GoUnder/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

type FofaConfig struct {
	Email string `json:"email"`
	Key   string `json:"key"`
}

type FofaResponse struct {
	Error   bool            `json:"error"`
	Results [][]string      `json:"-"`
	Msg     string          `json:"errmsg"`
	Raw     json.RawMessage `json:"results"`
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

	return fmt.Errorf("cannot unserialize results field: %s", string(aux.Results))
}

var targetURL string
var pattern string
var fofaCfg *FofaConfig
var logFlag bool

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

func cdnLookup(input string) [][]string {
	var err error
	fofaCfg, err = loadFofaConfig()
	if err != nil {
		log.Fatalf("Error loading fofa config: %v\n", err)
	}

	patterns := []string{"host", "title", "icon"}
	if pattern != "" {
		patterns = []string{pattern}
	}

	resultSet := make([][]string, 0)

	for _, p := range patterns {
		queries, encoded := get_queries(p, input)
		if queries != nil {
			fmt.Printf("[+] Query string loaded: %s   %s\n", queries[0][:len(queries[0])-len(utils.FofaRules())], "+ <Fofa filter cdn Rules>...")
		}
		for _, enc := range encoded {
			for _, ip := range Query(enc, "ip,port,host,org,country,region,city") {
				if len(ip) > 0 {
					resultSet = append(resultSet, ip)
				}
			}
		}
	}
	if len(resultSet) > 0 {
		resultSet = unique2D(resultSet)
		fmt.Println("\n✅ Promising target(s) found: ")

		var logContent strings.Builder
		for _, ip := range resultSet {
			line := strings.Join(ip, ", ")
			fmt.Println("-", line)
			if logFlag {
				logContent.WriteString(line + "\n")
			}
		}

		// --- 新增日志记录逻辑 ---
		if logFlag {
			saveToLog(input, logContent.String())
		}
		// -----------------------

		return resultSet
	} else {
		fmt.Println("\n❌ Could not find possible IP.")
		return nil
	}
}

// 去重 [][]string
func unique2D(input [][]string) [][]string {
	seen := make(map[string]bool)
	var result [][]string
	for _, arr := range input {
		key := strings.Join(arr, ",")
		if !seen[key] {
			seen[key] = true
			result = append(result, arr)
		}
	}
	return result
}

func get_queries(p string, input string) ([]string, []string) {
	var queries, encodedQueries []string

	switch p {
	case "host":
		q := fmt.Sprintf(`host="%s" `, extractHost(input)) + utils.FofaRules()
		queries = append(queries, q)

	case "title":
		titles, _ := get_titles(input)
		for _, title := range titles {
			fmt.Println("[+] Get website title:", title)
			q := fmt.Sprintf(`title="%s" `, title) + utils.FofaRules()
			queries = append(queries, q)
		}

	case "icon":
		iconHash, err := getFaviconHash(input)
		if err != nil {
			fmt.Println("get icon_hash failed:", err)
			break
		}
		fmt.Println("[+] Favicon hash loaded:", iconHash)
		q := fmt.Sprintf(`icon_hash="%s" `, iconHash) + utils.FofaRules()
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
		trimmed := strings.TrimSpace(strings.Join(title, ""))
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
		return nil, fmt.Errorf("cannot get valid website title")
	}

	return titles, nil
}

// Fetches and parses the HTML to find icon links
func extractIconLinks(baseURL string) ([]string, error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var icons []string
	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return icons, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.Data == "link" {
				var href, rel string
				for _, a := range t.Attr {
					if strings.ToLower(a.Key) == "rel" {
						rel = strings.ToLower(a.Val)
					}
					if strings.ToLower(a.Key) == "href" {
						href = a.Val
					}
				}
				if strings.Contains(rel, "icon") && href != "" {
					icons = append(icons, href)
				}
			}
		}
	}
}

// Convert relative URLs to absolute
func resolveURL(base string, ref string) string {
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	baseParsed, err := url.Parse(base)
	if err != nil {
		return ref
	}
	return baseParsed.ResolveReference(u).String()
}

// Check if icon URL returns 200 OK
func isValidURL(iconURL string) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Head(iconURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Main function to find the best valid icon
func GetValidFaviconURL(pageURL string) (string, error) {
	icons, err := extractIconLinks(pageURL)
	if err != nil {
		return "", err
	}

	// Add default /favicon.ico to candidates
	icons = append(icons, "/favicon.ico")

	checked := make(map[string]bool)

	for _, icon := range icons {
		iconURL := resolveURL(pageURL, icon)
		if checked[iconURL] {
			continue
		}
		checked[iconURL] = true

		if isValidURL(iconURL) {
			return iconURL, nil
		}
	}

	return "", fmt.Errorf("no valid favicon found")
}
func getFaviconHash(input string) (string, error) {
	// host := extractHost(input)
	url := input
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	// 下载 favicon
	favURL, err := GetValidFaviconURL(url)
	if err != nil {
		return "", err
	}
	hash, err := utils.GetIconHashFromURL(favURL)
	if err != nil {
		return "", err
	}
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
	configDir := "configs"
	filename := "fofa.json"
	path := filepath.Join(configDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			sysType := runtime.GOOS
			switch sysType {
			case "linux":
				configDir = filepath.Join(os.Getenv("HOME"), ".config", "GoUnder")
			case "windows":
				configDir = filepath.Join(os.Getenv("APPDATA"), "GoUnder")
			case "darwin":
				configDir = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "GoUnder")
			}
			path := filepath.Join(configDir, filename)
			data, err = os.ReadFile(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					log.Printf("Config file not found: %s\n", path)
					if err := os.MkdirAll(configDir, 0755); err != nil {
						return nil, fmt.Errorf("creating config failed: %w", err)
					}
					// 默认配置
					defaultCfg := FofaConfig{Email: "", Key: ""}
					defaultData, _ := json.MarshalIndent(defaultCfg, "", "  ")

					// 写入文件
					if err := os.WriteFile(path, defaultData, 0644); err != nil {
						return nil, fmt.Errorf("writing config file failed: %w", err)

					}
				}
				log.Printf("Config file created: %s\n❗ Please complete the config file: %s", path, path)
				return nil, err
			}
		}
	}
	err = json.Unmarshal(data, &fofaCfg)
	if fofaCfg.Email == "" || fofaCfg.Key == "" {
		log.Println("❗ Please complete the fofa config file with your email and API key.")
		os.Exit(1)
	}
	fmt.Printf("[+] Fofa account config loaded: %s\n", fofaCfg.Email)
	return fofaCfg, err
}

func Query(encodedQuery string, fields ...string) [][]string {
	client := resty.New()
	var result FofaResponse
	f := ""
	if len(fields) > 0 {
		f = fields[0]
	}

	_, err := client.R().
		SetQueryParams(map[string]string{
			"email":   fofaCfg.Email,
			"key":     fofaCfg.Key,
			"qbase64": encodedQuery,
			"size":    "100",
			"fields":  f,
		}).
		SetResult(&result).
		Get("https://fofa.info/api/v1/search/all")

	if err != nil {
		fmt.Println("request FOFA API failed:", err)
		fmt.Println(result)
		return nil
	}

	if result.Error {
		fmt.Printf("FOFA return error: %s\n", result.Msg)
		return nil
	}

	results := make([][]string, 0)
	for _, entry := range result.Results {
		if len(entry) > 0 && entry[0] != "" {
			results = append(results, entry)
		}
	}
	return results

	// var unique []string
	// for ip := range results {
	// 	unique = append(unique, ip)
	// }
	// return unique
}

func saveToLog(input string, content string) {
	// 1. 提取主机名作为文件名
	host := extractHost(input)
	host = strings.ReplaceAll(host, ":", "_") // 防止 Windows 下端口号导致的文件名非法

	// 2. 创建 logs 目录
	logDir := "logs"
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		_ = os.MkdirAll(logDir, 0755)
	}

	// 3. 构造完整路径 (例如: logs/example.com.log)
	fileName := filepath.Join(logDir, host+".log")

	// 4. 以追加模式打开文件，如果不存在则创建
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("⚠️  Failed to write log: %v\n", err)
		return
	}
	defer f.Close()

	// 5. 写入时间戳和内容
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("--- Scan at %s ---\n%s\n", timestamp, content)

	if _, err := f.WriteString(logEntry); err == nil {
		fmt.Printf("\n📝 Results appended to: %s\n", fileName)
	}
}

func init() {
	cdnCmd.Flags().StringVarP(&targetURL, "url", "u", "", "targetURL, eg:https://example.com")
	cdnCmd.Flags().StringVarP(&pattern, "pattern", "p", "", "[host | title | icon], default: all")
	cdnCmd.Flags().BoolVarP(&logFlag, "log", "", true, "log the details")
	rootCmd.AddCommand(cdnCmd)
}
