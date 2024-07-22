package rpc

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
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

func (c *FesHttpClient) Get(url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.responseHandle(request)
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
