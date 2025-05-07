package test

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-resty/resty/v2"
)

func main() {
	client := resty.New()
	param := map[string]string{"k": "v"}
	resp, err := client.R().
		SetQueryParams(param).
		SetHeader("Accept", "application/json").
		Get("https://httpbin.org/get")

	if err != nil {
		log.Fatal(err)
	}
	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		log.Fatalf("JSON解析失败: %v", err)
	}

	// 访问已知字段
	fmt.Printf("请求来源IP: %v\n", result["origin"])
	fmt.Printf("请求URL: %v\n", result["url"])
	for k, v := range result["headers"].(map[string]interface{}) {
		fmt.Println(k + ": " + v.(string))
	}

	fmt.Println(resp.String())
}
