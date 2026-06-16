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
	"encoding/json"
	"net/http"
)

// LoginHandler 测试抖音云免登录调用
// 小程序通过 cloud.callContainer 请求云托管服务时，抖音云会在 HTTP header 中注入用户身份信息，
// 服务端无需 code2session 即可直接获取，匿名登录仅能拿到 X-TT-ANONYMOUS-OPENID。
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	openID := r.Header.Get("X-TT-OPENID")
	if openID == "" || len(openID) == 0 {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"err_no":  -1,
			"err_msg": "X-TT-OPENID 为空，请通过小程序 cloud.callContainer 调用",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"err_no":  0,
		"err_msg": "success",
		"data": map[string]string{
			"openid":           openID,
			"appid":            r.Header.Get("X-TT-APPID"),
			"unionid":          r.Header.Get("X-TT-UNIONID"),
			"anonymous_openid": r.Header.Get("X-TT-ANONYMOUS-OPENID"),
			"env":              r.Header.Get("X-TT-ENV"),
			"source":           r.Header.Get("X-TT-SOURCE"),
			"client_ip":        r.Header.Get("X-Forwarded-For"),
		},
	})
}
