package Network

import (
	"encoding/json"
	"fmt"
	"github.com/davyxu/golog"
	"github.com/mitchellh/mapstructure"
	"io/ioutil"
	"github.com/Blizzardx/GoGameServer/Core/Common"
	"net/http"
	"reflect"
	"time"
)

// 用户端处理
type HttpEventCallback func(path string, msg interface{}) interface{}
type httpMessageInfo struct {
	Callback    HttpEventCallback
	MessageType reflect.Type
}
type httpServer struct {
	log      *golog.Logger
	handlers map[string]*httpMessageInfo
	port     string
	certFile string
	keyFile  string
}

func ListenAtHttp() *httpServer {
	httpServer := &httpServer{}
	httpServer.init()

	return httpServer
}
func (server *httpServer) Start(port string) {
	server.port = port
	mux := http.NewServeMux()
	for path := range server.handlers {
		mux.HandleFunc(path, server.onHttpHandler)
	}

	httpServer := &http.Server{
		Addr:           server.port,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}

	server.log.Debugln("begin listen http at port ", server.port)

	if server.certFile != "" && server.keyFile != "" {
		err := httpServer.ListenAndServeTLS(server.certFile, server.keyFile)
		if err != nil {
			server.log.Errorln("error on ListenAndServeTLS: ", err)
		} else {
			server.log.Infoln("listen https at port ", server.port, server.certFile, server.keyFile)
		}
	} else {
		err := httpServer.ListenAndServe()
		if err != nil {
			server.log.Errorln("error on ListenAndServe: ", err)
		} else {
			server.log.Infoln("listen http at port ", server.port)
		}
	}
}
func (server *httpServer) SetCert(certFile, keyFile string) {
	server.certFile = certFile
	server.keyFile = keyFile
}

//注册消息处理器
func (server *httpServer) RegisterMessage(path string, messageType reflect.Type, callback HttpEventCallback) {
	//path 必须 / 开头 ，path 不能重复注册
	server.handlers[path] = &httpMessageInfo{Callback: callback, MessageType: messageType}
}
func (server *httpServer) init() {
	server.log = golog.New("core.httpServer")
	server.handlers = map[string]*httpMessageInfo{}
}
func (server *httpServer) onHttpHandler(w http.ResponseWriter, req *http.Request) {
	Common.SafeCall(func() {
		if handler, ok := server.handlers[req.URL.Path]; ok {
			req.ParseForm()
			w.Header().Set("Access-Control-Allow-Origin", "*")
			server.log.Debugln("http received url ", req.URL)
			//decode message
			msg, err := server.decodeMsg(req.Form, handler.MessageType)
			if err != nil {
				server.log.Infoln("解析http请求参数失败", req.URL, err)
				return
			}
			server.log.Infoln("解析http请求参数", req.URL, msg)
			responseMsg := handler.Callback(req.URL.Path, msg)
			respMsg, err := server.encodeMsg(responseMsg)
			if err != nil {
				server.log.Errorln("序列化resp消息时发生错误", err)
				return
			}

			server.log.Debugln("http response ", string(respMsg))
			w.Write(respMsg)
		}
	})
}
func (server *httpServer) encodeMsg(msgBody interface{}) ([]byte, error) {
	return json.Marshal(msgBody)
}
func (server *httpServer) decodeMsg(form map[string][]string, msgType reflect.Type) (interface{}, error) {
	// create message body
	parserForm := map[string]string{}
	for k, v := range form {
		if len(v) > 0 {
			parserForm[k] = v[0]
		}
	}
	msgBody := reflect.New(msgType).Interface()
	err := mapstructure.Decode(parserForm, msgBody)
	if err != nil {
		return nil, err
	}
	return msgBody, nil
}

//http Get
func HttpGet(url string, result interface{}) error {
	log.Debugln("send request to url ", url)
	client := &http.Client{}
	response, err := client.Get(url)
	defer func() {
		if response != nil && response.Body != nil {
			response.Body.Close()
		}
	}()

	if err != nil {
		log.Errorln(err)
		return err
	}

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Errorln(err)
		return err
	}

	log.Debugln("receive response string ", string(bytes))

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		log.Errorln(err)
		return err
	}
	msg := fmt.Sprintf("%+v", result)
	log.Debugln("receive response ", msg)
	return nil
}
func HttpPost(url string) {

}
