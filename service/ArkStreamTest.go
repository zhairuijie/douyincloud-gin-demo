/*
Copyright (year) Bytedance Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ArkStreamHandler 流式调用大模型，以 SSE 形式向客户端转发增量内容
func ArkStreamHandler(w http.ResponseWriter, r *http.Request) {
	chat := r.URL.Query().Get("chat")
	if chat == "" {
		chat = "请介绍下字节豆包"
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"model": "ep-20260611121609-r9s46",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": chat,
			},
		},
		"stream": true,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", "https://ark-vg.dyc.ivolces.com/api/v3/chat/completions", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ark-90938081-59ee-46a8-b6b9-8d6dff6ea918-71bbe")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			jsonStr := line[6:]
			if jsonStr == "[DONE]" {
				break
			}
			var d arkStreamData
			if err := json.Unmarshal([]byte(jsonStr), &d); err != nil {
				continue
			}
			if len(d.Choices) == 0 {
				continue
			}
			if _, err := fmt.Fprintf(w, "%s", d.Choices[0].Delta.Content); err != nil {
				return
			}
			flusher.Flush()
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("扫描错误:", err)
	}
}

type arkStreamData struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"delta"`
		Index int `json:"index"`
	} `json:"choices"`
}
