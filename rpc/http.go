package rpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type FesHttpClient struct {
	client http.Client
}

func NewFesHttpClient() *FesHttpClient {
	client := http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       3 * time.Second,
	}
	return &FesHttpClient{
		client: client,
	}
}
func (c *FesHttpClient) Response(req *http.Request) ([]byte, error) {
	return c.responseHandle(req)
}

func (c *FesHttpClient) Get(url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *FesHttpClient) PostForm(url string, args map[string]any) ([]byte, error) {
	request, err := c.FormRequest(http.MethodPost, url, args)
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *FesHttpClient) PostJson(url string, args map[string]any) ([]byte, error) {
	request, err := c.JsonRequest(http.MethodPost, url, args)
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
}

func (c *FesHttpClient) GetRequest(method, url string, args map[string]any) (*http.Request, error) {
	return http.NewRequest(method, url, strings.NewReader(c.toValues(args)))
}

func (c *FesHttpClient) FormRequest(method, url string, args map[string]any) (*http.Request, error) {
	return http.NewRequest(method, url, strings.NewReader(c.toValues(args)))
}

func (c *FesHttpClient) JsonRequest(method, url string, args map[string]any) (*http.Request, error) {
	argsJson, _ := json.Marshal(args)
	return http.NewRequest(method, url, bytes.NewReader(argsJson))
}

func (c *FesHttpClient) responseHandle(request *http.Request) ([]byte, error) {
	resp, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status is %d", resp.StatusCode)
	}
	reader := bufio.NewReader(resp.Body)
	buf := make([]byte, 127)
	var body []byte
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF || n == 0 {
			break
		}
		body = append(body, buf[:n]...)
	}
	defer resp.Body.Close()
	return body, nil

}

func (c *FesHttpClient) toValues(args map[string]any) string {
	params := url.Values{}
	for k, v := range args {
		params.Set(k, fmt.Sprintf("%v", v))
	}
	return params.Encode()
}
