package integrations

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"io/ioutil"
	"net/http"
)

type SecureProxyClient struct {
	Url    string
	client *http.Client
}

func NewSecureProxy(url string) *SecureProxyClient {
	return &SecureProxyClient{
		Url:    url,
		client: &http.Client{},
	}
}

type secureProxyRequest struct {
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (p *SecureProxyClient) DoRequest(method, url string, headers map[string]string) ([]byte, error) {
	body := secureProxyRequest{
		Method:  method,
		Url:     url,
		Headers: headers,
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	res, err := p.client.Post(p.Url+"/proxy", "application/json", bytes.NewBuffer(encoded))
	if err != nil {
		sentry.Error(err)
		return nil, errors.New("error encoding request")
	}

	defer res.Body.Close()

	if errorHeader := res.Header.Get("x-proxy-error"); errorHeader != "" {
		return nil, errors.New(errorHeader)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("integration request returned status code %d", res.StatusCode)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}
