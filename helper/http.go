package helper

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/0829why/svrframe/logx"
)

var httpTransport *http.Transport

func init() {
	httpTransport = http.DefaultTransport.(*http.Transport)
}

func SetHttpTransport(t *http.Transport) {
	httpTransport = t
}

func DoPost(Domain string, headers map[string]string, data string) (resp *http.Response, response []byte, err error) {

	req, err := http.NewRequest("POST", Domain, strings.NewReader(data))
	if err != nil {
		logx.ErrorF("DoPost NewRequest err = %v", err)
		return
	}
	client := &http.Client{Timeout: 3 * time.Second}
	client.Transport = httpTransport
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	logx.DebugF("DoPost Domain = %s", Domain)
	logx.DebugF("DoPost data = %s", data)

	resp, err = client.Do(req)
	if err != nil {
		logx.ErrorF("DoPost client.Do err = %v", err)
		return
	}
	response, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logx.ErrorF("DoPost StatusCode = %d, status = %s, resp msg = %+v", resp.StatusCode, resp.Status, string(response))
	}
	return
}
func DoGet(Domain string, headers map[string]string, params map[string]string) (resp *http.Response, response []byte, err error) {

	Url, err := url.Parse(Domain)
	if err != nil {
		return
	}
	querys := url.Values{}
	for k, v := range params {
		querys.Set(k, v)
	}
	Url.RawQuery = querys.Encode()
	urlPath := Url.String()
	logx.DebugF("urlPath = %s", urlPath)

	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		logx.ErrorF("DoGet NewRequest err = %v", err)
		return
	}
	client := &http.Client{Timeout: 3 * time.Second}
	client.Transport = httpTransport
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err = client.Do(req)
	if err != nil {
		logx.ErrorF("DoGet client.Do err = %v", err)
		return
	}
	response, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logx.ErrorF("DoGet StatusCode = %d, status = %s, resp msg = %+v", resp.StatusCode, resp.Status, string(response))
	}
	return
}

func DoGetDirect(Domain string, headers map[string]string, query string) (resp *http.Response, response []byte, err error) {

	encode_query := query //url.QueryEscape(query)
	urlPath := Domain
	if len(encode_query) > 0 {
		urlPath += "?" + encode_query
	}
	logx.DebugF("urlPath = %s", urlPath)

	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		logx.ErrorF("DoGetDirect NewRequest err = %v", err)
		return
	}
	client := &http.Client{Timeout: 3 * time.Second}
	client.Transport = httpTransport
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err = client.Do(req)
	if err != nil {
		logx.ErrorF("DoGetDirect client.Do err = %v", err)
		return
	}
	response, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		logx.ErrorF("DoGetDirect StatusCode = %d, status = %s, resp msg = %+v", resp.StatusCode, resp.Status, string(response))
	}
	return
}
