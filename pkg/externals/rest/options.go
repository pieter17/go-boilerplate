package rest

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Option func(*Client)

type RoundTripWrapperFn func(parent http.RoundTripper) http.RoundTripper

func WithRoundTripWrapper(fn RoundTripWrapperFn) Option {
	return func(client *Client) {
		if client.httpClient.Transport == nil {
			client.httpClient.Transport = fn(http.DefaultTransport)
		} else {
			client.httpClient.Transport = fn(client.httpClient.Transport)
		}
	}
}

func WithDialer(dialer func(ctx context.Context, network, addr string) (net.Conn, error)) Option {
	return func(client *Client) {
		if dialer == nil {
			return
		}
		t, ok := client.httpClient.Transport.(*http.Transport)
		if !ok {
			return
		}
		t.DialContext = dialer
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.httpClient.Timeout = timeout
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}
