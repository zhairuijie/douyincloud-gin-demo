package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"time"
)

func Ark(ctx *gin.Context) {
	chat, ex := ctx.GetQuery("chat")
	if !ex {
		chat = "请介绍下字节豆包"
	}

	w := ctx.Writer
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	_, ok := w.(http.Flusher)
	if !ok {
		log.Panic("server not support") //浏览器不兼容
	}

	data := map[string]interface{}{
		"model": "ep-20240723163732-zr9qw",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": chat,
			},
		},
		"stream": true,
	}

	// 将数据编码为 JSON
	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}

	// 创建请求
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", "https://ark-vg.dyc.ivolces.com/api/v3/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer 3dad2a1f-7c1c-4350-80f8-dd057ffa3aba")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	// 处理流式响应
	scanner := bufio.NewScanner(resp.Body)
	// 自定义分隔符，这里以换行符为例
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			jsonStr := line[6:]
			var data Data
			err := json.Unmarshal([]byte(jsonStr), &data)
			if err != nil {
				continue
			}
			_, err = fmt.Fprintf(w, "%s", data.Choices[0].Delta.Content)
			if err != nil {
				continue
			}
			w.(http.Flusher).Flush()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("扫描错误:", err)
	}
}

type Data struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"delta"`
		Index int `json:"index"`
	} `json:"choices"`
}
