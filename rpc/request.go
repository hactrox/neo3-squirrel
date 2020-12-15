package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"neo3-squirrel/util/log"
	"strings"
	"sync"
	"time"

	eParser "github.com/go-errors/errors"
	"github.com/valyala/fasthttp"
)

var (
	client = &fasthttp.Client{
		MaxConnWaitTimeout: 10 * time.Second,
	}

	reqLock sync.RWMutex
)

type responseCommon struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func generateRequestBody(method string, params []interface{}) string {
	p := ""

	for _, param := range params {
		switch param := param.(type) {
		case int8, uint8,
			int16, uint16,
			int, uint,
			int32, uint32,
			int64, uint64:
			p += fmt.Sprintf("%d, ", param)
		case string:
			p += fmt.Sprintf("\"%s\", ", param)
		case []StackItem:
			parsed := ""
			for _, stackItem := range param {
				parsed += fmt.Sprintf(", {\"type\": \"%s\", \"value\": \"%s\"}", stackItem.Type, stackItem.Value)
			}
			p += fmt.Sprintf("[%s], ", parsed[2:])
		default:
			err := fmt.Errorf("the RPC parameter type must be integer or string. current type=%T, value=%v", param, param)
			panic(err)
		}
	}

	if p != "" {
		p = p[:len(p)-2]
	}

	body := `{
		"jsonrpc": "2.0",
		"method": "` + method + `",
		"params": [
			` + p + `
		],
		"id": 1
	}
	`
	return body
}

func request(minHeight uint, params string, target interface{}) {
	reqLock.RLock()
	// log.Debugf("rpc request: minHeight=%d, params=%s", minHeight, params)

	requestBody := []byte(params)
	resp := fasthttp.AcquireResponse()
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.SetBody(requestBody)

	for {
		url, ok, currBestHeight := selectNode(minHeight)
		if !ok {
			time.Sleep(50 * time.Millisecond)
			if allNodesDown || currBestHeight == -1 {
				reqLock.RUnlock()
				request(minHeight, params, target)
				return
			}

			reqLock.RUnlock()
			return
		}

		req.SetRequestURI(url)
		err := client.Do(req, resp)
		if err != nil {
			if !strings.Contains(err.Error(), "timed out") &&
				!strings.Contains(err.Error(), "connection refused") &&
				!strings.Contains(err.Error(), "the server closed connection before returning the first response byte.") {
				log.Error(err)
			}
			nodeUnavailable(url)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		break
	}

	bodyBytes := resp.Body()

	err := json.Unmarshal(bodyBytes, target)
	if err != nil {
		log.Error(errors.New(eParser.Wrap(err, 0).ErrorStack()))
		log.Errorf("Request body: %v", string(requestBody))
		log.Errorf("Response: %v", string(bodyBytes))
	}

	reqLock.RUnlock()
}
