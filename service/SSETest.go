package service

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "log"
    "net/http"
    "time"
)

func SSE(ctx *gin.Context) {
	w := ctx.Writer

	// sse 头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	_, ok := w.(http.Flusher)

	if !ok {
		log.Panic("server not support") //浏览器不兼容
	}
	for i := 0; i < 10; i++ {
		_, err := fmt.Fprintf(w, "data: %s\n\n", fmt.Sprintf("dsdf%d", i))
		if err != nil {
			return
		}
		w.(http.Flusher).Flush()
		time.Sleep(time.Second)
	}
}
