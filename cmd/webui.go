package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// ====== 嵌入静态文件 ======
//
//go:embed webui/static/*
var staticFiles embed.FS

// ====== 变量 ======
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
	fmt.Printf("[+] Starting web ui on %s:[%s]...\n", host, port)
	router := gin.Default()

	// 将嵌入的静态文件系统映射到 Gin
	subFS, err := fs.Sub(staticFiles, "webui/static")
	if err != nil {
		panic("加载静态文件失败: " + err.Error())
	}
	router.StaticFS("/static", http.FS(subFS))

	// 根路径返回 index.html
	router.GET("/", func(c *gin.Context) {
		file, err := staticFiles.ReadFile("webui/static/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error loading index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", file)
	})

	// API 分组
	api := router.Group("/api")
	{
		api.GET("/cdn", cdnHandler)
		api.GET("/fingerprint", fpHandler)
	}

	// 启动 HTTP 服务
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
	for _, parts := range cdnLookupResult {
		if len(parts) != 7 {
			continue
		}
		results = append(results, gin.H{
			"ip":      parts[0],
			"port":    parts[1],
			"host":    parts[2],
			"org":     parts[3],
			"country": parts[4],
			"region":  parts[5],
			"city":    parts[6],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"cdnData": results,
	})
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

	var results []gin.H
	original := fingerprintLookup(website, engine)
	for _, tech := range original {
		if len(tech) > 0 {
			results = append(results, gin.H{
				"tech":        tech["tech"],
				"version":     tech["version"],
				"description": tech["description"],
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"techData": results,
	})
}

func init() {
	webuiCmd.Flags().StringVarP(&port, "port", "p", "8080", "listening port, eg: 8080")
	webuiCmd.Flags().StringVarP(&host, "host", "a", "localhost", "host address, eg: localhost")
	rootCmd.AddCommand(webuiCmd)
}
