package test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-resty/resty/v2"
)

type WhatcmsConfig struct {
	Key string `json:"key"`
}

var cfg *WhatcmsConfig
var engine string

func main() {
	whatcmdAnalyze("r3tr0y.github.io")
}

func whatcmdAnalyze(url string) {
	cfg, err := loadWhatcmsConfig()
	if err != nil {
		fmt.Errorf("error loading whatcms config.\n")
	} else {
		fmt.Println("Whatcms account config loaded")
	}
	fmt.Println(*cfg)
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}
	client := resty.New()

	resp, err := client.R().SetHeader("Accept", "application/json").
		SetQueryParams(map[string]string{"key": cfg.Key, "url": url}).
		Get("https://whatcms.org/API/Tech")

	if err != nil {
		log.Fatalf("request failed: %v", err)

	} else {
		fmt.Println("\n✅ Website fingerprints found in whatcms: ")
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
}
func loadWhatcmsConfig() (*WhatcmsConfig, error) {
	path := filepath.Join("../configs", "whatcms.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &cfg)
	return cfg, err
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
