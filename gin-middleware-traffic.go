package UBT

import (
	"bytes"
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

func GinEsLog(ubt *UBT) gin.HandlerFunc {
	return func(c *gin.Context) {
		ginStart := time.Now()
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		requestId := c.GetHeader("requestId")
		if requestId == "" {
			UUID, _ := uuid.NewRandom()
			requestId = UUID.String()
		}

		var body []byte
		var err error

		if c.Request.Body != nil {
			body, err = ioutil.ReadAll(c.Request.Body)
			if err != nil {
				body = []byte{}
			}
		}

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
				PostData: string(body),
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

		// after request
		log.Print(responseTime)

		// access the status we are sending
		status := c.Writer.Status()
		log.Println(status)
	}
}
