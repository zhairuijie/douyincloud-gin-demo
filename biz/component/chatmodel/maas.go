package chatmodel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"douyin/avatar/serv_template_open/biz/model/schema"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/pkg/errors"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

var (
	defaultURL = "http://ark-vg.dyc.ivolces.com/api/v3/chat/completions"
	//defaultURL = "http://ark.cn-beijing.volces.com/api/v3/chat/completions"
)

type ChatModel struct {
	client *http.Client
}

type Config struct {
	HTTPClient          *http.Client `json:"-"`
	Timeout             time.Duration
	RetryTimes          int
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

func NewChatModel(_ context.Context, config *Config) *ChatModel {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost, // 默认值2，设置过小会导致每个host的time_wait变多
			IdleConnTimeout:     config.IdleConnTimeout,
		},
		Timeout: config.Timeout,
	}
	return &ChatModel{
		client: client,
	}
}

// ChatCompletionStream 火山方舟大模型调用--流式返回
func (cm *ChatModel) ChatCompletionStream(ctx context.Context, request *model.ChatCompletionRequest) (resp *http.Response, err error) {
	// 直接使用方舟的结构体，避免追更新、协议不兼容
	if !request.Stream {
		request.Stream = true // 本接口强制流式
	}
	hlog.CtxDebugf(ctx, "ChatCompletionStream...")
	resp, err = cm.do(ctx, request, nil, true)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ChatCompletion 火山方舟大模型调用--非流式返回
func (cm *ChatModel) ChatCompletion(ctx context.Context, request *model.ChatCompletionRequest) (ret *model.ChatCompletionResponse, apiError *model.APIError, err error) {
	// 直接使用方舟的结构体，避免追更新、协议不兼容
	resp, err := cm.do(ctx, request, nil, false)
	if err != nil {
		return nil, nil, err
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "[openapi.Plugin] ioutil.ReadAll failed")
	}

	log.Println(fmt.Sprintf("ChatCompletion respBody:%v, http code:%d", string(respBody), resp.StatusCode))
	// 异常返回
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = json.Unmarshal(respBody, &apiError)
		if err != nil {
			return nil, nil, errors.Wrap(err, "ChatCompletion respBody json.Unmarshal failed")
		}
		return nil, apiError, errors.New(string(respBody))
	}
	// 正常返回，将结果反射为
	err = json.Unmarshal(respBody, &ret)

	return ret, nil, err
}

func (cm *ChatModel) do(ctx context.Context, reqParam interface{}, headers map[string]string, stream bool) (*http.Response, error) {
	var (
		req *http.Request
		err error
	)

	reqBytes, err := json.Marshal(reqParam)
	if err != nil {
		return nil, errors.Wrap(err, "cm.do json.Marshal failed")
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, defaultURL, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequestWithContex failed")
	}

	// req.Header.Set("Authorization", "Bearer a88fa0dc-580f-4631-b8b5-c0243257a402")
	if stream {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")
	}

	// 添加header
	for k, v := range headers {
		if req.Header.Get(k) != "" {
			req.Header.Set(k, v)
		} else {
			req.Header.Add(k, v)
		}
	}

	// 发起请求
	resp, err := cm.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "httpClient.Do failed")
	}

	return resp, nil
}

// SSE 消息会有两个结束状态 DONE（返回 DONE 这个字符串） 和 STOP（消息体中有 stop）
func ChatModel2ChunkData(ctx context.Context, bufData []byte) (chunk *schema.MaasChatCompletionStreamResponse, isStop bool, isDone bool, err error) {
	bufStr := string(bufData)
	data := strings.TrimPrefix(bufStr, "data: ")
	// 以下注释，最后一个空chunk不用返回
	if data == "[DONE]" {
		return nil, true, true, nil
	}
	chunk = &schema.MaasChatCompletionStreamResponse{}
	err = json.Unmarshal([]byte(data), chunk)
	if err != nil {
		return nil, false, false, errors.Wrap(err, "func MaasSse2ChunkData failed")
	}

	if chunk.Error != nil {
		hlog.CtxErrorf(ctx, "chunk data err:%v", chunk.Error)
		return chunk, true, false, errors.New(fmt.Sprintf("chunk err, code:%s, message:%s", chunk.Error.Code, chunk.Error.Message))
	}

	for _, choice := range chunk.Choices {
		// 正常结束
		if choice.FinishReason == "" {
			continue
		}

		if choice.FinishReason == "stop" {
			return chunk, true, false, nil
		}

		if choice.FinishReason != "" && choice.FinishReason != "null" {
			hlog.CtxWarnf(ctx, "choice.delta:%v, reason:%#v, reason:%v", choice.Delta, choice.FinishReason, choice.FinishReason)
			return chunk, true, false, errors.New(string(choice.FinishReason))
		}
	}

	return chunk, false, false, nil
}
