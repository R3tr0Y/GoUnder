package cmd

import (
	"fmt"
	"net/http"
	"os"
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
		api.GET("/hello", helloHandler)
		api.GET("/time", timeHandler)
		api.GET("/analyze", analyzeHandler)
	}
	server := &http.Server{
		Addr:         host + ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
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

func helloHandler(c *gin.Context) {
	name := c.DefaultQuery("name", "World")
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello, " + name + "!",
	})
}
func timeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"time": time.Now().Format(time.RFC3339),
	})
}
func analyzeHandler(c *gin.Context) {
	website := c.DefaultQuery("website", "")
	if website == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Website is required",
		})
		return
	}

	// 调用封装的函数获取真实 IP 和云服务
	realIP, cloudService := fmt.Println(website)

	// 调用封装的函数获取网站指纹
	architecture, middleware := fmt.Println(website)

	c.JSON(http.StatusOK, gin.H{
		"ip":           realIP,
		"cloudService": cloudService,
		"architecture": architecture,
		"middleware":   middleware,
	})
}

func init() {
	webuiCmd.Flags().StringVarP(&port, "port", "p", "8080", "listening port, eg: 8080")
	webuiCmd.Flags().StringVarP(&host, "host", "u", "localhost", "listening port, eg: 8080")
	rootCmd.AddCommand(webuiCmd)
}
