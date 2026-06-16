package service

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Login 测试抖音云免登录调用
// 小程序通过 cloud.callContainer 请求时，抖音云会在 HTTP header 中注入用户身份信息，
// 服务端无需 code2session 即可直接获取。
func Login(ctx *gin.Context) {
	openID := ctx.GetHeader("X-TT-OPENID")
	if openID == "" {
		ctx.JSON(http.StatusOK, gin.H{
			"err_no":  -1,
			"err_msg": "X-TT-OPENID 为空，请通过小程序 cloud.callContainer 调用",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"err_no":  0,
		"err_msg": "success",
		"data": gin.H{
			"openid":           openID,
			"appid":            ctx.GetHeader("X-TT-APPID"),
			"unionid":          ctx.GetHeader("X-TT-UNIONID"),
			"anonymous_openid": ctx.GetHeader("X-TT-ANONYMOUS-OPENID"),
			"env":              ctx.GetHeader("X-TT-ENV"),
			"source":           ctx.GetHeader("X-TT-SOURCE"),
			"client_ip":        ctx.GetHeader("X-Forwarded-For"),
		},
	})
}
