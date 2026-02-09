package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-resty/resty/v2"
	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/spf13/cobra"
)

var fingerprintCmd = &cobra.Command{
	Use:   "fingerprint",
	Short: "Analyze fingerprints of websites.",
	Run: func(cmd *cobra.Command, args []string) {
		if targetURL == "" {
			fmt.Println("❗ use -u for target URL")
			_ = cmd.Usage()
			os.Exit(1)
		}
		fingerprintLookup(targetURL, engine)
	},
}

func fingerprintLookup(url string, engine string) []map[string]string {
	switch engine {
	case "":
		fmt.Println("Automatically using wappalyzergo...")

		return wappalyzerAnalyze(url)
	case "wappalyzer":
		return wappalyzerAnalyze(url)
	case "whatcms":
		return whatcmdAnalyze(url)
	}
	return nil

}
func wappalyzerAnalyze(url string) []map[string]string { //[]map[string]string
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body) // Ignoring error for example

	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		log.Fatal(err)
		return nil
	} else {
		fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)
		if len(fingerprints) > 0 {
			var logContent strings.Builder
			fmt.Println("\n✅ Website fingerprints found in local wappalyzer database:")
			results := convertTechMap(fingerprints)
			for fingerprint := range fingerprints {
				fmt.Printf("- %v\n", fingerprint)
				if logFlag {
					logContent.WriteString(fingerprint + "\n")
				}
			}
			// --- 新增日志记录逻辑 ---
			if logFlag {
				saveToLog(url, logContent.String())
			}
			// -----------------------
			return results
		} else {
			fmt.Println("❌ No website fingerprints found!")
			return nil
		}

	}
}

func whatcmdAnalyze(url string) []map[string]string {
	outcome := []map[string]string{}
	whatcmsCfg, err := loadWhatcmsConfig()
	if err != nil {
		log.Fatalf("error loading whatcms config.\n")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	client := resty.New()

	resp, err := client.R().SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{"key": whatcmsCfg.Key, "url": url}).
		Get("https://whatcms.org/API/Tech")

	if err != nil {
		log.Fatalf("request whatcms failed: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		log.Fatalf("parsing JSON format failed: %v", err)
	}
	// 提取 results 字段
	results, ok := result["results"].([]interface{})
	if !ok {
		log.Fatalf("results field is not type of []interface{}")
	}
	// 遍历 results 并输出格式化信息
	if len(results) > 0 {
		var logContent strings.Builder
		fmt.Println("\n✅ Website fingerprints found in whatcms: ")
		for _, item := range results {
			obj, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			name := getString(obj["name"])
			version := getString(obj["version"])
			categories := joinCategories(obj["categories"])

			// 构建输出字符串
			output := "- " + name
			if version != "" {
				output += ": " + version
			}
			if categories != "" {
				if version != "" {
					output += ", " + categories
				} else {
					output += ", " + categories
				}
			}
			outcome = append(outcome, map[string]string{"tech": name, "version": version, "description": categories})
			fmt.Println(output)
			if logFlag {
				logContent.WriteString(output + "\n")
			}
		}
		// --- 新增日志记录逻辑 ---
		if logFlag {
			saveToLog(url, logContent.String())
		}
		// -----------------------
		return outcome
	} else {
		fmt.Println("❌ No website fingerprints found!")
		return nil
	}

}
func loadWhatcmsConfig() (*WhatcmsConfig, error) {
	// configDir := "configs"
	// fileName := "whatcms.json"
	// path := filepath.Join(configDir, fileName)
	// data, err := os.ReadFile(path)
	// if err != nil {
	// 	if errors.Is(err, os.ErrNotExist) {
	// 		// 如果文件不存在，创建目录和文件
	// 		log.Println("Config file not found, creating config file...")

	// 		// 确保目录存在
	// 		if err := os.MkdirAll(configDir, 0755); err != nil {
	// 			return nil, fmt.Errorf("creating config failed: %w", err)
	// 		}

	// 		// 默认配置
	// 		defaultCfg := WhatcmsConfig{Key: ""}
	// 		defaultData, _ := json.MarshalIndent(defaultCfg, "", "  ")

	// 		// 写入文件
	// 		if err := os.WriteFile(path, defaultData, 0644); err != nil {
	// 			return nil, fmt.Errorf("writing config file failed: %w", err)
	// 		}

	// 		log.Printf("Config file created: %s\nPlease complete the config file", path)
	// 		os.Exit(1)
	// 	}
	// 	return nil, err
	// }
	// err = json.Unmarshal(data, &whatcmsCfg)
	// return whatcmsCfg, err
	configDir := "configs"
	filename := "whatcms.json"
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
					defaultCfg := WhatcmsConfig{Key: ""}
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
	err = json.Unmarshal(data, &whatcmsCfg)
	if whatcmsCfg.Key == "" {
		log.Println("❗ Please complete the whatcms config file with your API key.")
		os.Exit(1)
	}
	fmt.Printf("[+] Whatcms account config loaded: %s***\n", whatcmsCfg.Key[:5])
	return whatcmsCfg, err
}

// 辅助函数：安全获取字符串
func getString(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// 辅助函数：拼接 categories
func joinCategories(val interface{}) string {
	arr, ok := val.([]interface{})
	if !ok {
		return ""
	}
	var cats []string
	for _, cat := range arr {
		if s, ok := cat.(string); ok {
			cats = append(cats, s)
		}
	}
	return strings.Join(cats, " ")
}

func convertTechMap(input map[string]struct{}) []map[string]string {
	result := []map[string]string{}

	for key := range input {
		entry := map[string]string{
			"tech":        "",
			"version":     "",
			"description": "",
		}

		// 判断是否包含冒号（表示版本信息）
		if parts := strings.SplitN(key, ":", 2); len(parts) == 2 {
			entry["tech"] = strings.TrimSpace(parts[0])
			entry["version"] = strings.TrimSpace(parts[1])
		} else {
			entry["tech"] = strings.TrimSpace(key)
		}
		result = append(result, entry)
	}
	return result
}

func init() {
	fingerprintCmd.Flags().StringVarP(&targetURL, "url", "u", "", "targetURL, eg: https://example.com")
	fingerprintCmd.Flags().StringVarP(&engine, "engine", "e", "", "engine for analyzing website fingerprints, [ wappalyzer | whatcms | ], default: wappalyzer")
	fingerprintCmd.Flags().BoolVarP(&logFlag, "log", "", true, "log the scan results")
	rootCmd.AddCommand(fingerprintCmd)
}
