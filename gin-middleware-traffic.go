package UBT

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// 获取post方法提交的body数据
func getPostData(c *gin.Context) string {
	var buf []byte
	var err error

	if c.Request.Body != nil {
		buf, err = ioutil.ReadAll(c.Request.Body)
		// 重置Body，不然下次无法读取
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
		if err != nil {
			buf = []byte{}
		}
	}

	return string(buf)
}

// 获取全部headers
func getHeaders(c *gin.Context, ignoreHeaders []string) map[string][]string {
	reqHeaders := map[string][]string{}
	for k, v := range c.Request.Header {
		if _, ok := find(ignoreHeaders, k); !ok {
			reqHeaders[k] = v
		}
	}
	return reqHeaders
}
func getHeadersAsString(c *gin.Context, ignoreHeaders []string) string {
	reqHeaders := getHeaders(c, ignoreHeaders)
	s, err := json.Marshal(reqHeaders)
	if err != nil {
		return ""
	}
	return string(s)
}

// 获取或者生成请求id。如果客户端传了就使用客户端的。
func getRequestId(c *gin.Context) string {
	requestId := c.GetHeader("requestId")
	if requestId == "" {
		UUID, _ := uuid.NewRandom()
		requestId = UUID.String()
	}
	return requestId
}

func GinEsLog(ubt *UBT) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginStart := time.Now()
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		requestId := getRequestId(c)

		ubt.base(&Message{
			LogLevel: INFO,
			Msg: "request",
			Request: &ReqMessage{
				Url:      c.FullPath(),
				Method:   c.Request.Method,
				Path:     c.Request.URL.Path,
				Ip:       c.ClientIP(),
				Query:    c.Request.URL.RawQuery,
				Id:       requestId,
				Api:      c.GetHeader("method"), // 这个是调用接口的method，不是http method
				Ua:       c.GetHeader("user-agent"),
				PostData: getPostData(c),
				Headers:  getHeadersAsString(c, []string{}),
			},
		}, &ExtraMessage{
			LogType: "request",
		})

		c.Next()

		responseTime := time.Now().Sub(ginStart).Milliseconds()
		ubt.base(&Message{
			LogLevel: INFO,
			Msg: "response",
			Response: &ResMessage{
				ResponseTime: strconv.FormatInt(responseTime, 10),
				Data:         blw.body.String(),
			},
		}, &ExtraMessage{LogType: "response"})

		// access the status we are sending
		status := c.Writer.Status()
		log.Println(status)
	}
}
