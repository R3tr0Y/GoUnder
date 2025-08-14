package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func FofaRules() string {
	cloudfrontFilter, err := GetCloudFrontFOFAFilter()
	if err != nil {
		cloudfrontFilter = `header!="cloudfront"`
	}
	cloudflareFilter, err := GetCloudflareFOFAFilter()
	if err != nil {
		cloudflareFilter = `header!="cloudflare"`
	}
	return `&& server!="cloudflare" && server!="alicdn" && server!="qcloud"` +
		` && server!="yunjiasu" && server!="yupaicloud" && cloud_name!="Cloudflare"` +
		` && server!="upyun" && server!="ws" && server!="cdnws"` +
		` && server!="china cache" && server!="fastly" && server!="akamai"` +
		` && server!="akamaighost" && server!="hwcdn" && server!="Byte-nginx"` +
		` && server!="wangzhansheshi" && server!="360wzws" && server!="incapsula"` +
		` && server!="stackpath" && server!="keycdn" && cloud_name!="cloudfront"` +
		` && org!="CLOUDFLARENET" && server!="layun.com" && server!="*cdn*" &&` +
		cloudfrontFilter + ` && ` +
		cloudflareFilter
}
func getCacheFilePath() (string, error) {
	var baseDir string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			// fallback to home dir
			baseDir = filepath.Join(homeDir, "AppData", "Roaming", "GoUnder")
		} else {
			baseDir = filepath.Join(appData, "GoUnder")
		}
	case "darwin":
		baseDir = filepath.Join(homeDir, "Library", "Application Support", "GoUnder")
	default: // linux 和其他类 unix
		// 你也可以用 XDG_CONFIG_HOME 环境变量，默认 ~/.config/GoUnder
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			baseDir = filepath.Join(homeDir, ".config", "GoUnder")
		} else {
			baseDir = filepath.Join(xdgConfig, "GoUnder")
		}
	}

	// 确保目录存在
	err = os.MkdirAll(baseDir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(baseDir, "cloudfront_ips_cache.json"), nil
}

const (
	CloudFrontCacheFile     = "cloudfront_ips_cache.json"
	CloudFrontCacheDays     = 30
	CloudFrontAPIURL        = "https://d7uri8nf7uskq.cloudfront.net/tools/list-cloudfront-ips"
	CloudflareCacheFileName = "cloudflare_ips_cache.json"
	CloudflareCacheDays     = 30
	CloudflareIPV4URL       = "https://www.cloudflare-cn.com/ips-v4/"
)

type CloudFrontCache struct {
	CreateTime time.Time `json:"create_time"`
	IPList     []string  `json:"ip_list"`
}
type CloudflareCache struct {
	CreateTime time.Time `json:"create_time"`
	IPList     []string  `json:"ip_list"`
}

type CloudFrontAPIResponse struct {
	IPv4Prefixes []string `json:"CLOUDFRONT_GLOBAL_IP_LIST"`
	IPv6Prefixes []string `json:"CLOUDFRONT_GLOBAL_IP_LIST_IPV6"`
}

// GetCloudFrontFOFAFilter 获取 CloudFront FOFA 过滤规则
func GetCloudFrontFOFAFilter() (string, error) {
	ipList, err := getCloudFrontIPs()
	if err != nil {
		return "", err
	}
	return buildFOFAFilter(ipList), nil
}

// 获取 CloudFront IP 列表（带缓存）
func getCloudFrontIPs() ([]string, error) {
	cacheFile, err := getCacheFilePath()
	if err != nil {
		return nil, err
	}

	// 读取缓存
	if data, err := os.ReadFile(cacheFile); err == nil {
		var cache CloudFrontCache
		if json.Unmarshal(data, &cache) == nil {
			if time.Since(cache.CreateTime).Hours() < CloudFrontCacheDays*24 {
				return cache.IPList, nil
			}
		}
	}

	// 访问 CloudFront 官方 IP 列表
	resp, err := http.Get(CloudFrontAPIURL)
	if err != nil {
		return nil, fmt.Errorf("下载 CloudFront IP 列表失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp CloudFrontAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析 CloudFront JSON 失败: %v", err)
	}

	ipList := apiResp.IPv4Prefixes

	// 写入缓存
	cache := CloudFrontCache{
		CreateTime: time.Now(),
		IPList:     ipList,
	}
	// 保存缓存
	cacheJSON, _ := json.MarshalIndent(cache, "", "  ")
	_ = os.WriteFile(cacheFile, cacheJSON, 0644)

	return ipList, nil
}

// 构建 FOFA 查询语句
func buildFOFAFilter(ipList []string) string {
	var filters []string
	filters = append(filters, `header!="cloudfront"`)
	for _, ip := range ipList {
		filters = append(filters, fmt.Sprintf(`ip!="%s"`, ip))
	}
	return strings.Join(filters, " && ")
}

// GetCloudflareFOFAFilter 获取 Cloudflare FOFA 过滤规则
func GetCloudflareFOFAFilter() (string, error) {
	ipList, err := getCloudflareIPs()
	if err != nil {
		return "", err
	}
	return buildFOFAFilter(ipList), nil
}

// 获取 Cloudflare IP 列表（带缓存）
func getCloudflareIPs() ([]string, error) {
	cacheFile, err := getCacheFilePathFor("cloudflare_ips_cache.json")
	if err != nil {
		return nil, err
	}

	// 读取缓存
	if data, err := os.ReadFile(cacheFile); err == nil {
		var cache CloudflareCache
		if json.Unmarshal(data, &cache) == nil {
			if time.Since(cache.CreateTime).Hours() < CloudflareCacheDays*24 {
				return cache.IPList, nil
			}
		}
	}

	// 访问 Cloudflare IPv4 列表
	resp, err := http.Get(CloudflareIPV4URL)
	if err != nil {
		return nil, fmt.Errorf("下载 Cloudflare IPv4 列表失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP 错误: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	var ipList []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ipList = append(ipList, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 写入缓存
	cache := CloudflareCache{
		CreateTime: time.Now(),
		IPList:     ipList,
	}
	cacheJSON, _ := json.MarshalIndent(cache, "", "  ")
	_ = os.WriteFile(cacheFile, cacheJSON, 0644)

	return ipList, nil
}

// getCacheFilePathFor 按系统获取指定缓存文件路径，复用之前的目录规则
func getCacheFilePathFor(filename string) (string, error) {
	var baseDir string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			baseDir = filepath.Join(homeDir, "AppData", "Roaming", "GoUnder")
		} else {
			baseDir = filepath.Join(appData, "GoUnder")
		}
	case "darwin":
		baseDir = filepath.Join(homeDir, "Library", "Application Support", "GoUnder")
	default:
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			baseDir = filepath.Join(homeDir, ".config", "GoUnder")
		} else {
			baseDir = filepath.Join(xdgConfig, "GoUnder")
		}
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(baseDir, filename), nil
}
