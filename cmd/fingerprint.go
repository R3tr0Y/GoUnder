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

func fingerprintLookup(url string, engine string) {
	switch engine {
	case "":
		fmt.Println("Automatically using wappalyzergo...")
		wappalyzerAnalyze(url)
	case "local":
		wappalyzerAnalyze(url)
	case "whatcms":
		whatcmdAnalyze(url)
	}

}
func wappalyzerAnalyze(url string) {
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
	} else {
		fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)
		if len(fingerprints) > 0 {
			fmt.Println("\n✅ Website fingerprints found in local database:")
			for fingerprint := range fingerprints {
				fmt.Printf("- %v\n", fingerprint)
			}
		} else {
			fmt.Println("❌ No website fingerprints found!")
		}

	}
}

type WhatcmsConfig struct {
	Key string `json:"key"`
}

var whatcmsCfg *WhatcmsConfig
var engine string

func whatcmdAnalyze(url string) {
	whatcmsCfg, err := loadWhatcmsConfig()
	if err != nil {
		log.Fatalf("error loading whatcms config.\n")
	} else {
		fmt.Println("Whatcms account config loaded: " + whatcmsCfg.Key[:5] + "***")
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

			fmt.Println(output)
		}
	} else {
		fmt.Println("❌ No website fingerprints found!")
	}

}
func loadWhatcmsConfig() (*WhatcmsConfig, error) {
	configDir := "configs"
	fileName := "whatcms.json"
	path := filepath.Join(configDir, fileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// 如果文件不存在，创建目录和文件
			log.Println("Config file not found, creating config file...")

			// 确保目录存在
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

			log.Printf("Config file created: %s\nPlease complete the config file", path)
			os.Exit(1)
		}
		return nil, err
	}
	err = json.Unmarshal(data, &whatcmsCfg)
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

func init() {
	fingerprintCmd.Flags().StringVarP(&targetURL, "url", "u", "", "targetURL, eg: https://example.com")
	fingerprintCmd.Flags().StringVarP(&engine, "engine", "e", "", "engine for analyzing website fingerprints, [local (wappalyzergo) | whatcms | ], default: local")
	rootCmd.AddCommand(fingerprintCmd)
}
