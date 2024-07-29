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
	"douyincloud-gin-demo/component"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func Hello(ctx *gin.Context) {
	target := ctx.Query("target")
	if target == "" {
		Failure(ctx, fmt.Errorf("param invalid"))
		return
	}
	fmt.Printf("target= %s\n", target)
	hello, err := component.GetComponent(target)
	if err != nil {
		Failure(ctx, fmt.Errorf("param invalid"))
		return
	}

	name, err := hello.GetName(ctx, "name")
	if err != nil {
		Failure(ctx, err)
		return
	}
	Success(ctx, name)
}

func SetName(ctx *gin.Context) {
	var req SetNameReq
	err := ctx.Bind(&req)
	if err != nil {
		Failure(ctx, err)
		return
	}
	hello, err := component.GetComponent(req.Target)
	if err != nil {
		Failure(ctx, fmt.Errorf("param invalid"))
		return
	}
	err = hello.SetName(ctx, "name", req.Name)
	if err != nil {
		Failure(ctx, err)
		return
	}
	Success(ctx, "")
}

func Failure(ctx *gin.Context, err error) {
	resp := &Resp{
		ErrNo:  -1,
		ErrMsg: err.Error(),
	}
	ctx.JSON(200, resp)
}

func Success(ctx *gin.Context, data string) {
	resp := &Resp{
		ErrNo:  0,
		ErrMsg: "success",
		Data:   data,
	}
	ctx.JSON(200, resp)
}

type HelloResp struct {
	ErrNo  int    `json:"err_no"`
	ErrMsg string `json:"err_msg"`
	Data   string `json:"data"`
}

type SetNameReq struct {
	Target string `json:"target"`
	Name   string `json:"name"`
}

type Resp struct {
	ErrNo  int         `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

// OpenAPI 直接调用openapi
func OpenAPI(w http.ResponseWriter, r *http.Request) {
	url1 := "http://developer.toutiao.com/api/apps/v2/token"
	method1 := "POST"

	payload1 := strings.NewReader(`{"secret":"56ac324f4b081369b1975d254e7cf832650afb50",
"grant_type":"client_credential",
"appid":"tt5daf2b12c2857910"}`)

	client1 := &http.Client{}
	req1, err := http.NewRequest(method1, url1, payload1)

	if err != nil {
		fmt.Println(err)
		return
	}
	req1.Header.Add("Content-Type", "text/plain")

	res1, err := client1.Do(req1)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res1.Body.Close()

	body1, _ := ioutil.ReadAll(res1.Body)
	var resp Resp2
	err = json.Unmarshal(body1, &resp)
	fmt.Println(err)
	token := resp.Data.AccessToken

	url := "http://developer.toutiao.com/api/apps/qrcode"
	method := "POST"

	payload := strings.NewReader(`{
    "access_token": "` + token + `",
    "appname": "douyin"
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
	data := make(map[string]string)
	data["请求结果(20位)"] = string(body)[0:20]
	data["结论判断"] = "请求结果为PNG开头乱码即为【通过】，如果有token相关报错视为【不通过】"

	//以下为设置返回（勿动）
	msg, err := json.Marshal(data)
	if err != nil {
		fmt.Fprint(w, "内部错误")
		return
	}
	w.Write(msg)
}

type Resp2 struct {
	ErrNo   int    `json:"err_no"`
	ErrTips string `json:"err_tips"`
	Data    struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ExpiresAt   int    `json:"expiresAt"`
	} `json:"data"`
}

// PingHandler 火山校验
func PingHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(45 * time.Millisecond)
	fmt.Fprintf(w, "pong!\n")
}

//aibot调用豆包
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
