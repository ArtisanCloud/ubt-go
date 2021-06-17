package UBT

import (
	"encoding/json"
	"fmt"
	"github.com/guonaihong/gout"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
	"strings"
)

type Message struct {
	AppName        string      `json:"appName,omitempty"`
	AppVersion     string      `json:"appVersion,omitempty"`
	ServerHostname string      `json:"serverHostname,omitempty"`
	ServerAddr     string      `json:"serverAddr,omitempty"`
	LogLevel       string      `json:"logLevel,omitempty"`
	Request        *ReqMessage `json:"req,omitempty"`
	Response       *ResMessage `json:"res,omitempty"`
	Msg            string      `json:"msg,omitempty"`
	Error          *ErrorMsg   `json:"error,omitempty"`
	ExtraMessage
	baseInfo
}

type ReqMessage struct {
	Ip       string `json:"ip,omitempty"`       // 客户端的ip
	Method   string `json:"method,omitempty"`   // 客户端请求方法。 "GET" | "POST"
	Path     string `json:"path,omitempty"`     // 客户端请求路径
	PostData string `json:"postData,omitempty"` // 客户端提交的参数，url query参数和data参数都在里面
	Query    string `json:"query,omitempty"`    // 客户端查询的get query参数
	Url      string `json:"url,omitempty"`      // 完整的url链接地址
	Id       string `json:"id,omitempty"`
	Api      string `json:"api,omitempty"`
	Ua       string `json:"ua,omitempty"`
	Headers  string `json:"headers,omitempty"`
}

type ResMessage struct {
	ContentType  string `json:"contentType,omitempty"`
	ResponseTime string `json:"responseTime,omitempty"`
	Data         string `json:"data,omitempty"`
}

type ExtraMessage struct {
	BusinessInfo map[string]string `json:"businessInfo,omitempty"`
	Module       string            `json:"module,omitempty"` // 模块名字。 例如：预约里面的支付模块埋点，就写成booking.payment或者booking.pay。
	LogType      string            `json:"logType,omitempty"`
}

type baseInfo struct {
	SdkVersion string `json:"sdkVersion"` // GO SDK版本
}

type UBT struct {
	options *ClientOptions

	hostname string
	addr     string

	resText           string // ci模式 返回的字符串。
	err               error  // ci专用
	messageTextStacks []string
	messageText       string // 要发送的消息内容，这个只是为了测试。
}

type ClientOptions struct {
	UBTServer     string
	AppName       string
	AppVersion    string
	IgnoreHeaders []string
	systemInfo    map[string]string
	DebugMode     bool // debug模式
	ci            bool // ci模式
}

func Init(options *ClientOptions) *UBT {
	return &UBT{
		options: options,
	}
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
	Code   int    `json:"code,omitempty"`   // 错误code
	Stacks string `json:"stacks,omitempty"` // 错误堆栈
	File   string `json:"file,omitempty"`   // 错误文件
	Line   string `json:"line,omitempty"`   // 错误行
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
		Msg:      err.Error(),
		Error:    errMsg,
	}, extra)
}

func (ubt *UBT) clear() {
	ubt.messageText = ""
	ubt.messageTextStacks = []string{}
}

func (ubt *UBT) base(m *Message, extra *ExtraMessage) {
	fn := func() {
		if ubt.options.ci {
			ubt.err = nil
		}

		// 追加自定义业务字段和module字段
		if extra != nil {
			m.BusinessInfo = extra.BusinessInfo
			m.Module = extra.Module
			m.LogType = extra.LogType
		}

		// 追加默认字段
		m.AppName = ubt.options.AppName
		m.AppVersion = ubt.options.AppVersion
		m.SdkVersion = SdkVersion
		m.ServerHostname, m.ServerAddr = ubt.getHostNameAddr()

		messageText, err := json.Marshal(m)
		if err != nil {
			return
		}

		if ubt.options.ci {
			ubt.messageText = string(messageText)
			ubt.messageTextStacks = append(ubt.messageTextStacks, ubt.messageText)
		}

		err = gout.
			POST(ubt.options.UBTServer + "/logging/v2").
			SetJSON(m).
			Debug(ubt.options.DebugMode).
			BindBody(&ubt.resText).
			Do()
		if err != nil {
			log.Println(err)
		}
		if ubt.options.ci {
			ubt.err = err
		}
	}

	// 在ci环境下面走同步，因为需要测试完成情况
	if ubt.options.ci {
		fn()
	} else {
		go fn()
	}
}

func (ubt *UBT) getHostNameAddr() (string, string) {
	if ubt.options.ci {
		return "hostname-xxx", "192.168.1.1"
	}
	if ubt.hostname != "" && ubt.addr != "" {
		return ubt.hostname, ubt.addr
	}
	hostname, _ := os.Hostname()
	addrs, _ := net.LookupIP(hostname)

	addrStr := ""
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			fmt.Println("IPv4: ", ipv4)
			addrStr = addrStr + " " + ipv4.String()
		}
	}

	return hostname, strings.Trim(addrStr, " ")
}
