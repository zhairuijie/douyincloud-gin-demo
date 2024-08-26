package service

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/pipiguanli/douyincloud_mock/consts"
	Err "github.com/pipiguanli/douyincloud_mock/errors"
	"github.com/pipiguanli/douyincloud_mock/utils"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
)

type WebhookQaExtra struct {
	QaPath           *string `json:"webhook_qa_path"`
	WebhookSignature *string `json:"webhookheader_x_douyin_signature"`
	WebhookMsgId     *string `json:"webhookheader_msg_id"`
	Logid            *string `json:"logid"`
}

func WebhookCallback(ctx *gin.Context) {
	reqPath := ctx.FullPath()

	// 请求头
	webhookSignature := ctx.Request.Header.Get(consts.WebhookHeader_X_Douyin_Signature)
	webhookMsgId := ctx.Request.Header.Get(consts.WebhookHeader_Msg_Id)
	logid := ctx.Request.Header.Get(consts.Header_X_TT_Logid)
	qaExtra := &WebhookQaExtra{
		QaPath:           &reqPath,
		WebhookSignature: &webhookSignature,
		WebhookMsgId:     &webhookMsgId,
		Logid:            &logid,
	}
	//if err := utils.CheckHeaders(ctx); err != nil {
	//	TemplateFailure(ctx, Err.NewQaError(Err.InvalidParamErr, err.Error()))
	//	return
	//}

	//if len(utils.GetHeaderByName(ctx, consts.Header_StressTag)) > 0 {
	//	// sleep 随机 100ms ~ 1000ms（0.1s ~ 0.5s）
	//	num := utils.GenerateRandInt(100, 500)
	//	time.Sleep(time.Duration(num) * time.Millisecond)
	//}

	// 请求体
	var reqBodyString string
	if ctx.Request.Body != nil {
		reqBodyBytes, _ := ioutil.ReadAll(ctx.Request.Body)
		reqBodyString = string(reqBodyBytes)
	}
	reqEvent := gjson.Get(reqBodyString, "event").String()
	log.Printf("[QA] 请求体request=%+v ,req.event=%+v", reqBodyString, reqEvent)

	switch reqEvent {
	case "verify_webhook":
		// 反序列化
		req := &ReqVerifyWebhook{}
		if err := json.Unmarshal([]byte(reqBodyString), req); err != nil {
			TemplateFailure(ctx, Err.NewQaError(Err.ParamsResolveErr, "reqBodyString反序列化异常", err.Error()))
			return
		}
		resp := &RespVerifyWebhook{
			Challenge: func() *int64 {
				if req.Content != nil {
					return req.Content.Challenge
				}
				return nil
			}(),
			QaExtra: qaExtra,
		}
		httpStatusCode := 200
		ctx.JSON(httpStatusCode, resp)
		log.Printf("[QA] 响应体response=%+v, httpStatusCode=%+v", utils.ToJsonString(resp), httpStatusCode)

	default:
		httpStatusCode := 200
		resp := &WebhookCallbackCommonResp{}
		ctx.JSON(httpStatusCode, resp)
		log.Printf("[QA] 响应体response=%+v, httpStatusCode=%+v", utils.ToJsonString(resp), httpStatusCode)
	}
}

type ContentVerifyWebhook struct {
	Challenge *int64 `json:"challenge"`
}

type ReqVerifyWebhook struct {
	Event      *string               `json:"event"`
	ClientKey  *string               `json:"client_key"`
	FromUserId *string               `json:"from_user_id"`
	ToUserId   *string               `json:"to_user_id"`
	Content    *ContentVerifyWebhook `json:"content"`
}
type RespVerifyWebhook struct {
	Challenge *int64          `json:"challenge"`
	QaExtra   *WebhookQaExtra `json:"qa_extra"`
}

type WebhookCallbackCommonResp struct {
	ErrNo   int         `json:"err_no"`
	ErrTips string      `json:"err_tips"`
	Content interface{} `json:"content"`
	QaExtra *QaExtra    `json:"qa_extra"`
}
