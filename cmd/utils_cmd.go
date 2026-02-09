package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// cdn cmd definition
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

var targetURL string
var pattern string
var fofaCfg *FofaConfig
var logFlag bool

// fingerprint cmd definition

type WhatcmsConfig struct {
	Key string `json:"key"`
}

var whatcmsCfg *WhatcmsConfig
var engine string

// log function
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

func extractHost(raw string) string {
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		return raw
	}
	u, _ := url.Parse(raw)
	return u.Host
}
