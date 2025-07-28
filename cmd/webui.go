package cmd

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var port string
var host string
var webuiCmd = &cobra.Command{
	Use:   "webui",
	Short: "Start web ui.",
	Run: func(cmd *cobra.Command, args []string) {

		startWebui(host, port)
	},
}

func startWebui(host string, port string) {
	fmt.Printf("Starting web ui on %s:[%s]...", host, port)
	router := gin.Default()
	router.Static("/static", "./webui/static")
	router.StaticFile("/", "./webui/static/index.html")

	api := router.Group("/api")
	{
		api.GET("/cdn", cdnHandler)
		api.GET("/fingerprint", fpHandler)
	}
	server := &http.Server{
		Addr:         host + ":" + port,
		Handler:      router,
		ReadTimeout:  100 * time.Second,
		WriteTimeout: 100 * time.Second,
		IdleTimeout:  150 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("服务器启动失败: " + err.Error())
		}
	}()

	// 等待中断信号以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func cdnHandler(c *gin.Context) {
	website := c.DefaultQuery("website", "")
	if website == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Website is required",
		})
		return
	}
	pattern = c.DefaultQuery("p", "")

	// 调用封装的函数获取真实 IP 和云服务
	cdnLookupResult := cdnLookup(website)
	var results []gin.H
	for key := range cdnLookupResult {
		parts := strings.Split(key, ",")
		if len(parts) != 7 {
			continue // 跳过格式不正确的键
		}

		// 创建 JSON 结构
		jsonData := gin.H{
			"ip":      parts[0],
			"port":    parts[1],
			"host":    parts[2],
			"org":     parts[3],
			"country": parts[4],
			"region":  parts[5],
			"city":    parts[6],
		}

		results = append(results, jsonData)
	}
	// var host, org string
	// 调用封装的函数获取网站指纹
	// architecture, middleware := fmt.Println(website)

	c.JSON(http.StatusOK, gin.H{
		"cdnData": results,
	})
}

type TechInfo struct {
	Tech        string `json:"tech"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// 数据处理函数
func parseTechEntries(entries map[string]bool) []TechInfo {
	var result []TechInfo

	for entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		var tech, version, description string

		// 分离 description（逗号部分）
		parts := strings.SplitN(entry, ",", 2)
		main := strings.TrimSpace(parts[0])
		if len(parts) == 2 {
			description = strings.TrimSpace(parts[1])
		}

		// 分离 tech 和 version（冒号部分）
		techParts := strings.SplitN(main, ":", 2)
		tech = strings.TrimSpace(techParts[0])
		if len(techParts) == 2 {
			version = strings.TrimSpace(techParts[1])
		}

		result = append(result, TechInfo{
			Tech:        tech,
			Version:     version,
			Description: description,
		})
	}

	return result
}

func fpHandler(c *gin.Context) {
	website := c.DefaultQuery("website", "")
	if website == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Website is required",
		})
		return
	}
	engine = c.DefaultQuery("e", "")

	original := fingerprintLookup(website, engine)
	fmt.Println(original)
	parsed := parseTechEntries(original)

	// 返回标准化 JSON 格式
	c.JSON(http.StatusOK, gin.H{
		"techData": parsed,
	})

}

func init() {
	webuiCmd.Flags().StringVarP(&port, "port", "p", "8080", "listening port, eg: 8080")
	webuiCmd.Flags().StringVarP(&host, "host", "u", "localhost", "listening port, eg: 8080")
	rootCmd.AddCommand(webuiCmd)
}
