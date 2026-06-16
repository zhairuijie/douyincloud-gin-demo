package implements

import (
	"bufio"
	"context"
	"douyin/avatar/serv_template_open/biz/model/schema"
	"io"
	"net/http"
	"strings"
	"time"

	"douyin/avatar/serv_template_open/biz/component/chatmodel"
	"douyin/avatar/serv_template_open/biz/component/prompt"
	"douyin/avatar/serv_template_open/biz/consts"
	"douyin/avatar/serv_template_open/biz/model/avatar_serv"
	"douyin/avatar/serv_template_open/biz/model/common"
	"douyin/avatar/serv_template_open/biz/util"

	"douyin/avatar/serv_template_open/biz/component/memory"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	arkmodel "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
)

// ChatStream [POST] 流式对话接口
func ChatStream(ctx context.Context, c *app.RequestContext, req *avatar_serv.ChatStreamRequest) {
	// todo user
	// 非流式请求大模型
	cm := chatmodel.NewChatModel(ctx, &chatmodel.Config{})
	// 构造方舟请求
	chatReq := &arkmodel.ChatCompletionRequest{
		// Model:       "ep-20241028085559-598tx", //"ep-20241028195717-lhk5r",//"ep-20241028085559-598tx", // 需要修改为用户自己的enpoint_id
		Model:       "ep-20241028195717-lhk5r", //"ep-20241028085559-598tx", // 需要修改为用户自己的enpoint_id
		Messages:    prompt.FromMessages(req.Message, req.ChatContext.MessageContext),
		MaxTokens:   4000,
		Temperature: 0.7,
		TopP:        0.7,
		Stream:      true, // 非常重要，流式返回必填
	}

	streamRet, err := cm.ChatCompletionStream(ctx, chatReq)
	if err != nil {
		util.SendErrResp(ctx, c, -1, err)
		return
	}

	processResp(ctx, c, req.Message.Content.GetContent(), streamRet, req)

}

func processResp(ctx context.Context, c *app.RequestContext, query string, resp *http.Response, req *avatar_serv.ChatStreamRequest) {
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			hlog.CtxErrorf(ctx, "[processResp], resp body close failed, err=%v", err)
		}
	}(resp.Body)

	scanner := bufio.NewScanner(resp.Body)
	respWriter := util.NewResponseWriter(c, ctx, false)
	respWriter.SetIsWriteChunked(true)
	isSegQuery := strings.HasPrefix(query, "你好")
	isCardQuery := strings.HasPrefix(query, "卡片")
	isSingleCardQuery := strings.HasPrefix(query, "单卡混排")
	isMultiCardQuery := strings.HasPrefix(query, "多卡混排")
	isTimeoutQuery := strings.HasPrefix(query, "超时")
	isMemoryQuery := strings.HasPrefix(query, "记忆")

	if isTimeoutQuery {
		time.Sleep(time.Second * 11)
	}

	if isMemoryQuery {
		// 获取长期记忆
		memReq := &schema.MemoryRecallRequest{
			Params: &schema.MemoryRecallParams{
				Dsl:   nil,
				Query: volcengine.String(req.Message.Content.Content),
				Limit: volcengine.Int32(10),
			},
			RequestInfo: &schema.AvatarRequestInfo{
				BizID:         req.BizContext.BizID,
				TrafficSource: req.CommonContext.TrafficSource,
				OpenID:        req.CommonContext.UserInfo.OpenID,
				AvatarAppID:   req.CommonContext.AvatarInfo.AvatarAppID,
				TenantID:      volcengine.String("ebtest_1"),
				ProviderID:    volcengine.String("ebtest_1"),
			},
		}
		memContents, err := memory.RecallMemory(ctx, memReq)
		if err != nil {
			util.SendErrResp(ctx, c, -1, err)
			return
		}
		hlog.CtxInfof(ctx, "memContent:%v", memContents)
	}

	idx := 0
	for scanner.Scan() {
		idx = idx + 1
		rawBytes := scanner.Bytes()
		hlog.CtxDebugf(ctx, "raw bytes:%v", string(rawBytes))
		if len(rawBytes) == 0 {
			continue
		}
		// 标准方舟chunk
		chunk, isEnd, isDone, err := chatmodel.ChatModel2ChunkData(ctx, rawBytes)
		if err != nil {
			hlog.CtxErrorf(ctx, "[ChatStream] error %v, bytes: %v", err, string(rawBytes))
			util.SendErrResp(ctx, c, -1, err)
			return
		}
		if chunk == nil {
			hlog.CtxWarnf(ctx, "[ChatStream] read chunk is nil, raw data: %v", string(rawBytes))
			continue
		}

		chunkContent := ""
		if len(chunk.Choices) > 0 && chunk.Choices[0] != nil {
			chunkContent = chunk.Choices[0].Delta.Content
		}
		chunkResp := &avatar_serv.ChatStreamResponse{
			StreamFinish: false,
			Content: &avatar_serv.CopilotContent{
				Type:      common.ContentType_ContentType_TEXT,
				Content:   chunkContent,
				Role:      common.Role_Role_Assistant,
				SegFinish: false,
				SegType:   common.SegType_SegType_Answer,
			},
			TraceInfo: nil,
		}
		if idx%3 == 0 && isSegQuery {
			chunkResp.Content.SegFinish = true
		}

		respWriter.WriteHttpResponse(chunkResp)
		if (isSingleCardQuery || isMultiCardQuery) && idx == 5 {
			respWriter.WriteHttpResponse(&avatar_serv.ChatStreamResponse{
				StreamFinish: false,
				Content: &avatar_serv.CopilotContent{
					Type:      common.ContentType_ContentType_CARD,
					Content:   consts.XiaoheCard,
					Role:      common.Role_Role_Assistant,
					SegFinish: true,
					SegType:   common.SegType_SegType_Answer,
				},
				TraceInfo: nil,
				BaseResp:  nil,
			})
		}
		hlog.CtxInfof(ctx, "idxxxxxxxxxxx:%v", idx)
		if isMultiCardQuery && idx == 11 {
			respWriter.WriteHttpResponse(&avatar_serv.ChatStreamResponse{
				StreamFinish: false,
				Content: &avatar_serv.CopilotContent{
					Type:      common.ContentType_ContentType_CARD,
					Content:   consts.XiaoheCard,
					Role:      common.Role_Role_Assistant,
					SegFinish: true,
					SegType:   common.SegType_SegType_Answer,
				},
				TraceInfo: nil,
				BaseResp:  nil,
			})
		}
		if isDone || isEnd {
			if isCardQuery {
				respWriter.WriteHttpResponse(&avatar_serv.ChatStreamResponse{
					StreamFinish: false,
					Content: &avatar_serv.CopilotContent{
						Type:      common.ContentType_ContentType_CARD,
						Content:   consts.XiaoheCard,
						Role:      common.Role_Role_Assistant,
						SegFinish: true,
						SegType:   common.SegType_SegType_Answer,
					},
					TraceInfo: nil,
					BaseResp:  nil,
				})
			}

			endChunkResp := &avatar_serv.ChatStreamResponse{
				StreamFinish: true,
				Content: &avatar_serv.CopilotContent{
					Type:      common.ContentType_ContentType_TEXT,
					Content:   "",
					Role:      common.Role_Role_Assistant,
					SegFinish: true,
					SegType:   common.SegType_SegType_Answer,
				},
				TraceInfo: nil,
			}
			respWriter.WriteHttpResponse(endChunkResp)
			break
		}
	}

	respWriter.WriteChunkedEnd()
}
