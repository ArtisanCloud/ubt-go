package UBT

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	// 注入自动捕获网络数据的中间件
	ubt := InitUBT()
	r.Use(GinEsLog(ubt))

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.POST("/ping", func(c *gin.Context) {
		c.String(200, "post pong")
	})
	return r
}

func TestPingRoute(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestHTTPTrafficMiddleWareByGet(t *testing.T)  {
	router := setupRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	ubtReqMessageText := ubt.messageTextStacks[0]
	ubtResMessageText := ubt.messageTextStacks[1]

	assert.Contains(t, ubtReqMessageText, `"logType":"request"`)
	assert.Contains(t, ubtReqMessageText, `"method":"GET"`)
	assert.Contains(t, ubtReqMessageText, `"url":"`)

	assert.Contains(t, ubtResMessageText, `"data":"pong"`)
	assert.Contains(t, ubtResMessageText, `"responseTime":"`)

	ubt.clear()
}

func TestHTTPTrafficMiddleWareByPostFormData(t *testing.T)  {
	router := setupRouter()
	w := httptest.NewRecorder()
	params := url.Values{}
	params.Set("a", "1")
	params.Set("b", "2")
	req, _ := http.NewRequest("POST", "/ping", strings.NewReader(params.Encode()))
	router.ServeHTTP(w, req)

	ubtReqMessageText := ubt.messageTextStacks[0]
	ubtResMessageText := ubt.messageTextStacks[1]

	assert.Contains(t, ubtReqMessageText, `"method":"POST"`)
	assert.Contains(t, ubtReqMessageText, `"path":"/ping"`)
	assert.Contains(t, ubtReqMessageText, `"postData":"a=1\u0026b=2"`)
	assert.Contains(t, ubtResMessageText, `"data":"post pong"`)

	ubt.clear()
}

func TestHTTPTrafficMiddleWareByPostJSON(t *testing.T)  {
	router := setupRouter()
	w := httptest.NewRecorder()
	var jsonStr = []byte(`{"a": 1, "b": 2}`)
	req, _ := http.NewRequest("POST", "/ping", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	ubtReqMessageText := ubt.messageTextStacks[0]
	ubtResMessageText := ubt.messageTextStacks[1]

	assert.Contains(t, ubtReqMessageText, `"method":"POST"`)
	assert.Contains(t, ubtReqMessageText, `"path":"/ping"`)
	assert.Contains(t, ubtReqMessageText, `"postData":"{\"a\": 1, \"b\": 2}"`)
	assert.Contains(t, ubtResMessageText, `"data":"post pong"`)

	ubt.clear()
}
