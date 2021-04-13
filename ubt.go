package ubt

import (
	"encoding/json"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/pkg/errors"
	"log"
)

type Message struct {
	AppName        string `json:"appName,omitempty"`
	AppVersion     string `json:"appVersion,omitempty"`
	serverHostname string
	serverAddr     string
	LogLevel       string `json:"logLevel,omitempty"`
	Request        *struct {
		Ip        string `json:"ip,omitempty"`         // 客户端的ip
		Method    string `json:"method,omitempty"`     // 客户端请求方法。 "GET" | "POST"
		Path      string `json:"path,omitempty"`       // 客户端请求路径
		PostData  string `json:"post_data,omitempty"`  // 客户端提交的参数，url query参数和data参数都在里面
		Query     string `json:"query,omitempty"`      // 客户端查询的get query参数
		RequestId string `json:"request_id,omitempty"` // 请求id，优先读取客户端生成的值。
		Url       string `json:"url,omitempty"`        // 完整的url链接地址
	} `json:"request,omitempty"`
	Response *struct {
		Data string `json:"data,omitempty"` // 只有debug等级的时候才会显示
	} `json:"response,omitempty"`
	//Error *struct {
	//	Code   int    `json:"code,omitempty"`   // 错误code
	//	Stacks string `json:"stacks,omitempty"` // 错误堆栈
	//	File   string `json:"file,omitempty"`   // 错误文件
	//	Line   string `json:"line,omitempty"`   // 错误行
	//} `json:"error,omitempty"`
	Msg string `json:"msg,omitempty"`
	Error *ErrorMsg `json:"error,omitempty"`
	ExtraMessage
	baseInfo
}

type ExtraMessage struct {
	BusinessInfo map[string]string `json:"businessInfo,omitempty"`
	Module       string            `json:"module,omitempty"` // 模块名字。 例如：预约里面的支付模块埋点，就写成booking.payment或者booking.pay。
}

type baseInfo struct {
	SdkVersion string `json:"sdkVersion"` // GO SDK版本
}

type UBT struct {
	UBTServer   string
	AppName     string
	AppVersion  string
	systemInfo  map[string]string
	DebugMode   bool   // debug模式
	ci          bool   // ci模式
	resText     string // ci模式 返回的字符串。
	err         error  // ci专用
	messageText string // 要发送的消息内容，这个只是为了测试。
}

func (ubt *UBT) Debug(msg string, extra *ExtraMessage) {
	message := &Message{
		Msg:      msg,
		LogLevel: DEBUG,
	}
	ubt.base(message, extra)
}

func (ubt *UBT) Info(msg string, extra *ExtraMessage) {
	message := &Message{
		Msg:      msg,
		LogLevel: INFO,
	}
	ubt.base(message, extra)
}

func (ubt *UBT) Warn(msg string, extra *ExtraMessage) {
	message := &Message{
		Msg:      msg,
		LogLevel: WARN,
	}
	ubt.base(message, extra)
}

func (ubt *UBT) Error(msg string, extra *ExtraMessage) {
	message := &Message{
		Msg:      msg,
		LogLevel: ERROR,
	}
	ubt.base(message, extra)
}

func (ubt *UBT) Critical(msg string, extra *ExtraMessage) {
	message := &Message{
		Msg:      msg,
		LogLevel: CRITICAL,
	}
	ubt.base(message, extra)
}
func (ubt *UBT) Alert(msg string, extra *ExtraMessage) {
	message := &Message{
		Msg:      msg,
		LogLevel: ALERT,
	}
	ubt.base(message, extra)
}
func (ubt *UBT) Fatal(msg string, extra *ExtraMessage) {
	message := &Message{
		Msg:      msg,
		LogLevel: FATAL,
	}
	ubt.base(message, extra)
}

type ErrorMsg struct {
	Stacks string `json:"stacks"`
}

// SendError 错误发送。 如果需要捕获错误堆栈，那么则需要使用github.com/pkg/errors
func (ubt *UBT) SendError(err error, extra *ExtraMessage) {
	errMsg := &ErrorMsg{
		Stacks: "",
	}
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			errMsg.Stacks += fmt.Sprintf("%+s:%d\n", f, f)
		}
	}

	ubt.base(&Message{
		LogLevel: ERROR,
		Msg: err.Error(),
		Error: errMsg,
	}, extra)
}

func (ubt *UBT) base(message *Message, extra *ExtraMessage) {
	if ubt.ci {
		ubt.err = nil
	}

	// 追加自定义业务字段和module字段
	if extra != nil {
		message.BusinessInfo = extra.BusinessInfo
		message.Module = extra.Module
	}

	// 追加默认字段
	message.AppName = ubt.AppName
	message.AppVersion = ubt.AppVersion
	message.SdkVersion = SdkVersion

	messageText, err := json.Marshal(message)
	if err != nil {
		return
	}

	if ubt.ci {
		ubt.messageText = string(messageText)
	}

	err = gout.
		POST(ubt.UBTServer + "/logging/v2").
		SetJSON(message).
		Debug(ubt.DebugMode).
		BindBody(&ubt.resText).
		Do()
	if err != nil {
		log.Println(err)
	}
	if ubt.ci {
		ubt.err = err
	}
}