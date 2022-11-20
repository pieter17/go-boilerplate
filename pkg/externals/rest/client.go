package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

type JSONHTMLEscaper interface {
	JSONHTMLEscaper() bool
}

type Client struct {
	httpClient *http.Client
	baseUrl    string
	headers    map[string]string
}

func New(baseUrl string, opt ...Option) *Client {
	c := &Client{
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			Timeout: 2 * time.Minute,
		},
		strings.TrimRight(baseUrl, "/"),
		map[string]string{"Content-Type": "application/json"},
	}
	for _, o := range opt {
		o(c)
	}
	return c
}

func (c *Client) WithRequestHeaders(headers map[string]string) *Client {
	nc := &Client{
		httpClient: c.httpClient,
		baseUrl:    c.baseUrl,
		headers:    map[string]string{},
	}

	for k, v := range c.headers {
		nc.headers[k] = v
	}
	for k, v := range headers {
		nc.headers[k] = v
	}
	return nc
}

func (c *Client) Call(ctx context.Context, method string, path string, body interface{}, response interface{}) (err error) {
	var b io.Reader

	if body != nil {
		bb := &bytes.Buffer{}
		enc := json.NewEncoder(bb)
		if x, ok := body.(JSONHTMLEscaper); ok {
			enc.SetEscapeHTML(x.JSONHTMLEscaper())
		}
		err := enc.Encode(body)
		if err != nil {
			return err
		}
		bb.Truncate(bb.Len() - 1)
		b = bb
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseUrl+path, b)
	if err != nil {
		return err
	}
	req.Close = true
	req.Header.Set("Connection", "Close")

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	res, err := c.httpClient.Do(req)
	if res != nil && res.Body != nil {
		defer func(c io.Closer) {
			c.Close()
		}(res.Body)
	}

	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode > 300 {
		err := errors.New(res.Status)
		return err
	}

	return json.NewDecoder(res.Body).Decode(&response)
}
