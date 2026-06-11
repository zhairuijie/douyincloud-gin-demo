package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

var (
	client              *http.Client
	requestTimeout      = 300 * time.Second
	maxIdleConns        = 1000
	maxIdleConnsPerHost = 100
	idleConnTimeout     = 10 * time.Second
)

func InitClient() {
	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConns,
			MaxIdleConnsPerHost: maxIdleConnsPerHost, // 默认值2，设置过小会导致每个host的time_wait变多
			IdleConnTimeout:     idleConnTimeout,
		},
		Timeout: requestTimeout,
	}
}

func Do(ctx context.Context, url, method string, reqParam interface{}, headers map[string]string, stream bool) (*http.Response, error) {
	var (
		req *http.Request
		err error
	)

	if stream {
		req, err = newStreamRequest(ctx, url, method, reqParam, headers)
	} else {
		req, err = build(ctx, url, method, reqParam, headers)
	}

	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "http client.Do failed")
	}

	return resp, nil
}

func newStreamRequest(ctx context.Context, url, method string, reqParam interface{}, headers map[string]string) (*http.Request, error) {
	req, err := build(ctx, url, method, reqParam, headers)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	return req, nil
}

func build(ctx context.Context, urlStr, method string, reqParam interface{}, headers map[string]string) (*http.Request, error) {
	var (
		req      *http.Request
		err      error
		reqBytes []byte
	)

	if reqParam != nil {
		if method == http.MethodGet {
			urlValues := ToURLValues(reqParam)
			urlStr = fmt.Sprintf("%s&%s", urlStr, urlValues.Encode())
		} else {
			reqBytes, err = json.Marshal(reqParam)
			if err != nil {
				return nil, err
			}
		}
	}
	req, err = http.NewRequestWithContext(ctx, method, urlStr, bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, errors.Wrap(err, "http.NewRequestWithContex failed")
	}

	for k, v := range headers {
		if req.Header.Get(k) != "" {
			req.Header.Set(k, v)
		} else {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

func ToURLValues(i interface{}) (values url.Values) {
	values = url.Values{}
	iVal := reflect.ValueOf(i).Elem()
	typ := iVal.Type()

	for i := 0; i < iVal.NumField(); i++ {
		f := iVal.Field(i)
		if f.Interface() == nil {
			continue
		}
		// You ca use tags here...
		// tag := typ.Field(i).Tag.Get("tagname")
		// Convert each type into a string for the url.Values string map
		var v string
		switch f.Interface().(type) {
		case int, int8, int16, int32, int64:
			v = strconv.FormatInt(f.Int(), 10)
		case uint, uint8, uint16, uint32, uint64:
			v = strconv.FormatUint(f.Uint(), 10)
		case float32:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 32)
		case float64:
			v = strconv.FormatFloat(f.Float(), 'f', 4, 64)
		case []byte:
			v = string(f.Bytes())
		case string:
			v = f.String()
		case bool:
			v = fmt.Sprintf("%t", f.Bool())
		}
		values.Set(typ.Field(i).Tag.Get("json"), v)
	}

	return
}
