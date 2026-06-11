package implements

import (
	"context"
	"douyin/avatar/serv_template_open/biz/model/avatar_serv"
	"douyin/avatar/serv_template_open/biz/model/common"
	"douyin/avatar/serv_template_open/biz/util"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func OnBoarding(ctx context.Context, c *app.RequestContext, req *avatar_serv.OnBoardingRequest) {
	// todo user
	resp := []string{
		"你好,",
		"我是",
		"赵小小，",
		"一位女性情感专家，",
		"有什么可以帮助你？",
	}
	sugs := []string{
		"异地恋该如何长期维护稳定",
		"女朋友总是长期不回答消息，表明了什么",
		"如何给暗恋多年的女神表白",
	}

	respWriter := util.NewResponseWriter(c, ctx, false)
	respWriter.SetIsWriteChunked(true)
	for idx, c := range resp {
		r := &avatar_serv.OnBoardingResponse{
			StreamFinish: false,
			Content: &avatar_serv.CopilotContent{
				Type:      common.ContentType_ContentType_TEXT,
				Content:   c,
				Role:      common.Role_Role_Assistant,
				SegFinish: false,
				SegType:   common.SegType_SegType_Answer,
			},
			TraceInfo: &common.TraceInfo{TraceInfo: "{\"logid\":\"\"}"},
		}

		if idx == len(resp)-1 {
			r.Content.SegFinish = true // todo：需要这样标记么？
		}
		err := respWriter.WriteHttpResponse(r)
		if err != nil {
			hlog.CtxErrorf(ctx, "respWriter.WriteHttpResponse failed, err:%v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	for idx, s := range sugs {
		r := &avatar_serv.OnBoardingResponse{
			StreamFinish: false,
			Content: &avatar_serv.CopilotContent{
				Type:      common.ContentType_ContentType_TEXT,
				Content:   s,
				Role:      common.Role_Role_Assistant,
				SegFinish: false,
				SegType:   common.SegType_SegType_FollowUp,
			},
			TraceInfo: &common.TraceInfo{TraceInfo: "{\"logid\":\"\"}"},
		}

		if idx == len(sugs)-1 {
			r.Content.SegFinish = true
			r.StreamFinish = true
		}

		err := respWriter.WriteHttpResponse(r)
		if err != nil {
			hlog.CtxErrorf(ctx, "respWriter.WriteHttpResponse failed, err:%v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	respWriter.WriteChunkedEnd()
	return
}
