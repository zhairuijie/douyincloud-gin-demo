package dal

import (
	"douyin/avatar/serv_template_open/biz/dal/http"
)

func Init() {
	http.InitClient()
}
