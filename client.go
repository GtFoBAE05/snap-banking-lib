package snap

import (
	"fmt"
	"net/http"
	"snap-banking-lib/adapter"
	"snap-banking-lib/adapter/bca"
	"snap-banking-lib/adapter/bri"
	"snap-banking-lib/internal/httpclient"
	"snap-banking-lib/model"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Client interface {
	GetAdapter(model.BankCode) (adapter.Adapter, error)
}

type client struct {
	adapter        map[model.BankCode]adapter.Adapter
	config         model.Config
	httpClient     httpclient.HttpClient
	httpTimeout    time.Duration
	httpLogger     httpclient.Logger
	debug          bool
	curlLog        bool
	requestIdKey   string
	requestIdLabel string
	retryConfig    *httpclient.RetryConfig
	cbConfig       *httpclient.CircuitBreakerConfig
	tracerProvider trace.TracerProvider
}

type ClientOption func(*client) error

func NewClient(config model.Config, opts ...ClientOption) (Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	c := &client{
		adapter: make(map[model.BankCode]adapter.Adapter),
		config:  config,
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if c.httpClient == nil {
		httpOpts := []httpclient.Option{
			httpclient.WithDebug(c.debug),
			httpclient.WithCurlLog(c.curlLog),
		}
		if c.httpLogger != nil {
			httpOpts = append(httpOpts, httpclient.WithLogger(c.httpLogger))
		}
		if c.requestIdKey != "" {
			httpOpts = append(httpOpts, httpclient.WithRequestIdKey(c.requestIdKey, c.requestIdLabel))
		}

		if c.retryConfig != nil {
			httpOpts = append(httpOpts, httpclient.WithRetry(*c.retryConfig))
		}
		if c.cbConfig != nil {
			httpOpts = append(httpOpts, httpclient.WithCircuitBreaker(*c.cbConfig))
		}
		c.httpClient = httpclient.NewHttpClient(c.httpTimeout, httpOpts...)
	}

	tracerProvider := c.tracerProvider
	if tracerProvider == nil {
		tracerProvider = otel.GetTracerProvider()
	}
	tracer := tracerProvider.Tracer("snap-banking-lib")

	for bankCode, bankConfig := range config.Banks {
		var adp adapter.Adapter
		var err error

		switch model.BankCode(strings.ToUpper(string(bankCode))) {
		case model.BankBCA:
			adp, err = bca.NewAdapter(bankConfig, c.httpClient, tracer)
		case model.BankBRI:
			adp, err = bri.NewAdapter(bankConfig, c.httpClient, tracer)

		default:
			return nil, fmt.Errorf("unsupported bank: %s", bankCode)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create adapter for %s: %w", bankCode, err)
		}

		c.adapter[bankCode] = adp
	}

	return c, nil
}

func (c *client) GetAdapter(bankCode model.BankCode) (adapter.Adapter, error) {
	if !c.config.HasBank(bankCode) {
		return nil, fmt.Errorf("bank code %s is not initialized", bankCode)
	}

	return c.adapter[bankCode], nil
}

func WithHttpHandler(h httpclient.HttpHandler) ClientOption {
	return func(c *client) error {
		c.httpClient = httpclient.NewHttpClient(0, httpclient.WithHandler(h))
		return nil
	}
}

func WithStdHttpClient(h *http.Client) ClientOption {
	return func(c *client) error {
		c.httpClient = httpclient.NewHttpClient(0, httpclient.WithHandler(h))
		return nil
	}
}

func WithLogger(l httpclient.Logger) ClientOption {
	return func(c *client) error {
		c.httpLogger = l
		return nil
	}
}

func WithDebug() ClientOption {
	return func(c *client) error {
		c.debug = true
		return nil
	}
}

func WithCurlLog() ClientOption {
	return func(c *client) error {
		c.curlLog = true
		return nil
	}
}

func WithRequestId(key, label string) ClientOption {
	return func(c *client) error {
		c.requestIdKey = key
		c.requestIdLabel = label
		return nil
	}
}

func WithRetry(cfg httpclient.RetryConfig) ClientOption {
	return func(c *client) error {
		c.retryConfig = &cfg
		return nil
	}
}

func WithCircuitBreaker(cfg httpclient.CircuitBreakerConfig) ClientOption {
	return func(c *client) error {
		c.cbConfig = &cfg
		return nil
	}
}

func WithTracerProvider(tp trace.TracerProvider) ClientOption {
	return func(c *client) error {
		c.tracerProvider = tp
		return nil
	}
}
