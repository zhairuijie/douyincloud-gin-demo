package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzResp "github.com/cloudwego/hertz/pkg/protocol/http1/resp"
)

const (
	PredictTransSign = "\r\n"
	FinishTransSign  = "0\r\n\r\n"
)

var (
	defaultErrNO = -1 // 默认所错误位
	defaultNO    = 0
	defaultMsg   = "success"
)

type httpResp struct {
	ErrNO  int         `json:"err_no"`
	ErrMsg string      `json:"err_msg"`
	Data   interface{} `json:"data"`
}

func SendSucResp(ctx context.Context, c *app.RequestContext, data interface{}) {
	resp := &httpResp{
		ErrNO:  defaultNO,
		ErrMsg: defaultMsg,
		Data:   data,
	}

	c.JSON(http.StatusOK, resp)
}

func SendErrResp(ctx context.Context, c *app.RequestContext, code int, err error) {
	if err != nil && code == 0 {
		code = defaultErrNO
	}
	hlog.CtxErrorf(ctx, "code=%v, err=%v", code, err)
	ret := &httpResp{
		ErrNO:  code,
		ErrMsg: err.Error(),
		Data:   make(map[string]interface{}),
	}

	c.JSON(http.StatusOK, ret)
}

type ResponseWriter struct {
	c              *app.RequestContext
	ctx            context.Context
	hasWriteHeader bool
	isWriteChunked bool
	writeLock      sync.Mutex
}

func NewResponseWriter(c *app.RequestContext, ctx context.Context, isWriteChunked bool) *ResponseWriter {
	respWriter := &ResponseWriter{
		c:              c,
		ctx:            ctx,
		hasWriteHeader: false,
		isWriteChunked: isWriteChunked,
		writeLock:      sync.Mutex{},
	}
	if respWriter.isWriteChunked {
		respWriter.OpenHertzChunked()
	}
	return respWriter
}

func (respWriter *ResponseWriter) SetIsWriteChunked(isWriteChunked bool) {
	if isWriteChunked && !respWriter.isWriteChunked {
		respWriter.isWriteChunked = isWriteChunked
		respWriter.OpenHertzChunked()
	}
}

func (respWriter *ResponseWriter) OpenHertzChunked() {
	// 1.hertz chunked即时返回给端
	respWriter.c.Response.HijackWriter(hertzResp.NewChunkedBodyWriter(&respWriter.c.Response,
		respWriter.c.GetWriter()))
	// 2.声明为短连接，服务端发送完数据后立即close掉本次链接，否则默认为长链接
	respWriter.c.Response.Header.SetConnectionClose(true)
}

func (respWriter *ResponseWriter) WriteHttpResponse(data interface{}) error {
	resp := httpResp{
		ErrNO:  defaultNO,
		ErrMsg: defaultMsg,
		Data:   data,
	}
	byteResp, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	if respWriter.isWriteChunked {
		return respWriter.writeChunkedResponse(byteResp)
	} else {
		return respWriter.WriteNoChunkedResponse(byteResp)
	}
}

func (respWriter *ResponseWriter) WriteHttpErrorResponse(code int, err error) {
	if err != nil && code == 0 {
		code = defaultErrNO
	}
	hlog.CtxErrorf(respWriter.ctx, "code=%v, err=%v", code, err)

	ret := &httpResp{
		ErrNO:  code,
		ErrMsg: err.Error(),
		Data:   make(map[string]interface{}),
	}

	respWriter.c.JSON(http.StatusOK, ret)
}

func (respWriter *ResponseWriter) WriteNoChunkedResponse(data []byte) error {
	respWriter.c.Response.Header.Set("Content-Type", "application/json")
	respWriter.c.Response.Header.SetStatusCode(http.StatusOK)
	if _, err := respWriter.c.Write(data); err != nil {
		return err
	}
	return nil
}

func (respWriter *ResponseWriter) writeChunkedResponse(data []byte) error {
	hLen, succ := getHexString(respWriter.ctx, len(data))
	if !succ {
		return errors.New("write data len to res data failed")
	}
	// logs.CtxInfo(respWriter.ctx, "WriteResponse,data_len=%v", hLen)
	data = StringToBytes(hLen + PredictTransSign + BytesToString(data) + PredictTransSign)
	respWriter.writeLock.Lock()
	defer respWriter.writeLock.Unlock()
	if !respWriter.hasWriteHeader {
		respWriter.hasWriteHeader = true
		respWriter.writeChunkedHeader()
	}
	if _, err := respWriter.c.Write(data); err != nil {
		return err
	}
	if err := respWriter.c.Flush(); err != nil {
		return err
	}
	return nil
}

func (respWriter *ResponseWriter) writeChunkedHeader() {
	respWriter.c.Response.Header.Set("Transfer-Encoding", "chunked")
	respWriter.c.Response.Header.Set("X-Use-Chunk", "1")
	respWriter.c.Response.Header.Set("Content-Type", "application/json")
	respWriter.c.Response.SetStatusCode(http.StatusOK)
}

func (respWriter *ResponseWriter) WriteChunkedEnd() error {
	if !respWriter.isWriteChunked {
		return nil
	}
	respWriter.writeLock.Lock()
	defer respWriter.writeLock.Unlock()
	if _, err := respWriter.c.Write(StringToBytes(FinishTransSign)); err != nil {
		return err
	}
	if err := respWriter.c.Flush(); err != nil {
		return err
	}
	return nil
}

func getHexString(ctx context.Context, length int) (string, bool) {
	hLen := strconv.FormatInt(int64(length), 16)
	if len(hLen) > 6 {
		hlog.CtxFatalf(ctx, "send data error, len[%d] is too large or invalid", length)
		return "", false
	}
	hLen = fmt.Sprintf("%06s", hLen)
	return hLen, true
}
