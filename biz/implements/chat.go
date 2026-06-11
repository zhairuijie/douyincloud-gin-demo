package implements

import (
	"context"

	"douyin/avatar/serv_template_open/biz/component/chatmodel"
	"douyin/avatar/serv_template_open/biz/component/prompt"
	"douyin/avatar/serv_template_open/biz/model/avatar_serv"
	"douyin/avatar/serv_template_open/biz/model/base"
	"douyin/avatar/serv_template_open/biz/model/common"

	"github.com/cloudwego/hertz/pkg/app"
	arkmodel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

// Chat [POST]
func Chat(ctx context.Context, c *app.RequestContext, req *avatar_serv.ChatRequest) (resp *avatar_serv.ChatResponse) {
	// todo user
	// 非流式请求大模型
	cm := chatmodel.NewChatModel(ctx, &chatmodel.Config{})
	// 构造方舟请求
	chatReq := &arkmodel.ChatCompletionRequest{
		Model:       "endpoint_id", // todo：需要修改为用户自己的endpoint_id
		Messages:    prompt.FromMessages(req.Message, req.ChatContext.MessageContext),
		MaxTokens:   4000,
		Temperature: 0.7,
		TopP:        0.7,
		Stream:      false,
	}

	ret, _, err := cm.ChatCompletion(ctx, chatReq)
	if err != nil {
		resp = &avatar_serv.ChatResponse{
			Content:   nil,
			TraceInfo: nil,
			BaseResp: &base.BaseResp{
				StatusMessage: err.Error(),
				StatusCode:    -1,
				Extra:         nil,
			},
		}

		return resp
	}
	resp = &avatar_serv.ChatResponse{
		Content: []*avatar_serv.CopilotContent{
			{
				Type:      common.ContentType_ContentType_TEXT,
				Content:   volcengine.StringValue(ret.Choices[0].Message.Content.StringValue),
				Role:      common.Role_Role_Assistant,
				SegFinish: true,
				SegType:   common.SegType_SegType_Answer,
			},
		},
		TraceInfo: nil,
		BaseResp:  nil,
	}

	return resp
}
