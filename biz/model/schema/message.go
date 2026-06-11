package schema

import (
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type MaasChatCompletionStreamResponse struct {
	*model.ChatCompletionStreamResponse
	Error *model.APIError `json:"error"`
}

type MemoryRecallHttpResponse struct {
	Errno  int                   `json:"err_no"`
	ErrMsg string                `json:"err_msg"`
	LogID  string                `json:"log_id"`
	Data   *MemoryRecallResponse `json:"data"`
}
