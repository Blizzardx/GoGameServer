package main

import (
	"litgame.cn/Server/Core/Network"
	"reflect"
)

type LoginInfo struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}
type CommonResponse struct {
	ErrorCode int32  `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

func main() {
	httpServer := Network.ListenAtHttp()
	httpServer.RegisterMessage("/login", reflect.TypeOf(LoginInfo{}), httpMsgHandler)
	httpServer.Start(":8082")
}
func httpMsgHandler(path string, msg interface{}) interface{} {
	msgBody := msg.(*LoginInfo)

	return &CommonResponse{ErrorCode: 0, ErrorMsg: msgBody.UserName}
}
