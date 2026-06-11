package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	dalhttp "douyin/avatar/serv_template_open/biz/dal/http"
	"douyin/avatar/serv_template_open/biz/model/schema"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/pkg/errors"
)

const (
	// Host = "http://open-ai.byted.org/dy_open_api/avatar/atomic/api/"
	Host = "http://open-ai-byted-org.dyc.ivolces.com/dy_open_api/avatar/atomic/api/"
)

func RecallMemory(ctx context.Context, request *schema.MemoryRecallRequest) ([]string, error) {
	// 做http请求
	// func Do(ctx context.Context, url, method string, reqParam interface{}, headers map[string]string, stream bool) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", Host, "recall_memory/")
	headers := map[string]string{
		"x-tt-env":     "ppe_aiavatar_v1",
		"x-use-ppe":    "1",
		"Content-Type": "application/json",
		// "access-token": "clt.1506fe976005d3c453d781a25690d78c738ZZnSElho1Zmtx7DoABW20Qj5G_hl",
	}
	hlog.CtxInfof(ctx, "url:%v", url)
	resp, err := dalhttp.Do(ctx, url, http.MethodPost, request, headers, false)
	if err != nil {
		return nil, errors.Wrap(err, "http do failed")
	}

	// 解析resp
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "[RecallMemory] ioutil.ReadAll failed")
	}

	hlog.CtxInfof(ctx, "respBodyrespBody:%v", string(respBody))
	var memoryResp = &schema.MemoryRecallHttpResponse{}
	err = json.Unmarshal(respBody, memoryResp)
	if err != nil {
		return nil, errors.Wrap(err, "[CrawlV2] json.Unmarshal failed")
	}

	if memoryResp.Errno != 0 {
		return nil, errors.New(fmt.Sprintf("code=%d, msg=%s, log_id=%s", memoryResp.Errno, memoryResp.ErrMsg, memoryResp.LogID))
	}

	hlog.CtxInfof(ctx, "memoryResp:%v", memoryResp)
	if memoryResp.Data == nil || len(memoryResp.Data.Memories) == 0 {
		return nil, nil
	}

	var ret = []string{}
	for _, m := range memoryResp.Data.Memories {
		ret = append(ret, m.Content)
	}

	return ret, nil
}
