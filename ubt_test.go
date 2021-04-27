package UBT

import (
	"errors"
	"fmt"
	errors2 "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testAppName = "taishan-dev"
const testAppVersion = "0.1.0"

var ubt *UBT

func init()  {
	InitUBT()
}

func InitUBT() *UBT {
	if ubt != nil {
		return ubt
	}
	ubt = Init(&ClientOptions{
		UBTServer: "https://metrics-dev.spacecycle.cn",
		AppName: testAppName,
		AppVersion: testAppVersion,
		ci: true,
		DebugMode: true,
	})
	return ubt
}



func TestUBT_Debug(t *testing.T) {
	assertNew := assert.New(t)

	// 测试最简单的debug message
	ubt.Debug("hello, debug", nil)
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, debug",
			"logLevel":"debug"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")

	// 测试携带module参数
	ubt.Debug("hello, debug", &ExtraMessage{
		Module:       "ubt.debug",
	})
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, debug",
			"module": "ubt.debug",
			"logLevel":"debug"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")


	// 测试携带BusinessInfo参数
	ubt.Debug("hello, debug", &ExtraMessage{
		BusinessInfo: map[string]string{
			"test": "1",
			"test2": "1",
		},
		Module:       "ubt.debug",
	})
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, debug",
			"businessInfo": {
				"test": "1",
				"test2": "1"
			},
			"module": "ubt.debug",
			"logLevel":"debug"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")
}

func TestUBT_Info(t *testing.T) {
	assertNew := assert.New(t)

	// 测试最简单的debug message
	ubt.Info("hello, info", nil)
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, info",
			"logLevel":"info"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")
}

func TestUBT_Warn(t *testing.T) {
	assertNew := assert.New(t)

	// 测试最简单的debug message
	ubt.Warn("hello, warn", nil)
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, warn",
			"logLevel":"warn"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")
}

func TestUBT_Error(t *testing.T) {
	assertNew := assert.New(t)

	// 测试最简单的debug message
	ubt.Error("hello, error", nil)
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, error",
			"logLevel":"error"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")
}

func TestUBT_Critical(t *testing.T) {
	assertNew := assert.New(t)

	// 测试最简单的debug message
	ubt.Critical("hello, critical", nil)
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, critical",
			"logLevel":"critical"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")
}


func TestUBT_Alert(t *testing.T) {
	assertNew := assert.New(t)

	// 测试最简单的debug message
	ubt.Alert("hello, alert", nil)
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, alert",
			"logLevel":"alert"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")
}

func TestUBT_Fatal(t *testing.T) {
	assertNew := assert.New(t)

	// 测试最简单的debug message
	ubt.Fatal("hello, fatal", nil)
	assertNew.JSONEq(ubt.messageText, fmt.Sprintf(`
		{
			"appName": "%s",
			"appVersion":"%s",
			"sdkVersion": "%s",
			"msg": "hello, fatal",
			"logLevel":"fatal"
		}
	`, testAppName, testAppVersion, SdkVersion))
	assertNew.Equal(ubt.err, nil)
	assertNew.Contains(ubt.resText, "\"code\":0")
}

func fn1() error {
	return fn2()
}
func fn2() error {
	return fn3()
}
func fn3() error {
	err := errors2.New("test error2")
	return err
}

// 测试正常的Error
func TestUBT_SendError(t *testing.T) {
	assertNew := assert.New(t)
	err := errors.New("test error")
	ubt.SendError(err, nil)

	// 推断读取到了error的内容
	assertNew.Contains(ubt.messageText, "test error")

	assertNew.Contains(ubt.resText, "\"code\":0")
	assertNew.Equal(ubt.err, nil)
}

// 测试github.com/pkg/errors里面的Error类型
func TestUBT_SendError2(t *testing.T) {
	assertNew := assert.New(t)
	err2 := fn1()
	ubt.SendError(err2, nil)
	fmt.Println("test error finish")

	// 推断读取到了error的内容
	assertNew.Contains(ubt.messageText, "test error")

	assertNew.Contains(ubt.resText, "\"code\":0")
	assertNew.Equal(ubt.err, nil)
}

