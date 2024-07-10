package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TicketsBot/common/sentry"
	"io"
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
	Method   string            `json:"method"`
	Url      string            `json:"url"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     []byte            `json:"body,omitempty"`
	JsonBody json.RawMessage   `json:"json_body,omitempty"`
}

type requestBody interface {
	[]byte | any
}

func (p *SecureProxyClient) DoRequest(ctx context.Context, method, url string, headers map[string]string, bodyData requestBody) ([]byte, error) {
	body := secureProxyRequest{
		Method:  method,
		Url:     url,
		Headers: headers,
	}

	// nil will fall through anyway
	if bodyData != nil && (method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete) {
		switch v := bodyData.(type) {
		case []byte:
			base64.StdEncoding.Encode(body.Body, v)
		case any:
			encoded, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}

			body.JsonBody = json.RawMessage(encoded)
		}
	}

	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.Url+"/proxy", bytes.NewBuffer(encoded))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := p.client.Do(req)
	if err != nil {
		sentry.Error(err)
		return nil, errors.New("error proxying request")
	}

	defer res.Body.Close()

	if errorHeader := res.Header.Get("x-proxy-error"); errorHeader != "" {
		return nil, errors.New(errorHeader)
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("integration request returned status code %d", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return resBody, nil
}
