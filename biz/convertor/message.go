package convertor

import (
	"douyin/avatar/serv_template_open/biz/model/common"
)

func AvatarServRole2ArkRole(role common.Role) string {
	switch role {
	case common.Role_Role_System:
		return "system"
	case common.Role_Role_User:
		return "user"
	case common.Role_Role_Assistant:
		return "assistant"
	}

	return ""
}
