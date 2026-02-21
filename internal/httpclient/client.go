package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"moul.io/http2curl"
)

type HttpHandler interface {
	Do(*http.Request) (*http.Response, error)
}

type HttpClient interface {
	Do(ctx context.Context, method, url string, headers map[string]string, body []byte) (*http.Response, error)
}

type Logger interface {
	Debug(msg string, args ...any)
}

type defaultLogger struct{}

func (n *defaultLogger) Debug(_ string, _ ...any) {}

type RetryConfig struct {
	MaxAttempts uint
	Delay       time.Duration
	MaxDelay    time.Duration
}

type CircuitBreakerConfig struct {
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
	Threshold   uint32
}

type httpClient struct {
	handler        HttpHandler
	logger         Logger
	debug          bool
	curlLog        bool
	requestIdKey   string
	requestIdLabel string
	retry          *RetryConfig
	cb             *gobreaker.CircuitBreaker
	tracer         trace.Tracer
}

type Option func(*httpClient)

func WithLogger(l Logger) Option {
	return func(c *httpClient) { c.logger = l }
}

func WithDebug(enabled bool) Option {
	return func(c *httpClient) { c.debug = enabled }
}

func WithCurlLog(enabled bool) Option {
	return func(c *httpClient) { c.curlLog = enabled }
}

func WithHandler(h HttpHandler) Option {
	return func(c *httpClient) { c.handler = h }
}

func WithRequestIdKey(key, label string) Option {
	return func(c *httpClient) {
		c.requestIdKey = key
		c.requestIdLabel = label
	}
}

func WithRetry(cfg RetryConfig) Option {
	return func(c *httpClient) { c.retry = &cfg }
}

func WithCircuitBreaker(cfg CircuitBreakerConfig) Option {
	return func(c *httpClient) {
		c.cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
			MaxRequests: cfg.MaxRequests,
			Interval:    cfg.Interval,
			Timeout:     cfg.Timeout,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures >= cfg.Threshold
			},
		})
	}
}

func WithTracerProvider(tp trace.TracerProvider) Option {
	return func(c *httpClient) {
		c.tracer = tp.Tracer("snap-banking-lib/httpclient")
	}
}

func NewHttpClient(timeout time.Duration, opts ...Option) HttpClient {
	hc := &httpClient{
		handler: &http.Client{Timeout: timeout},
		logger:  &defaultLogger{},
		tracer:  otel.GetTracerProvider().Tracer("snap-banking-lib/httpclient"),
	}
	for _, opt := range opts {
		opt(hc)
	}
	return hc
}

func (c *httpClient) Do(ctx context.Context, method, url string, headers map[string]string, body []byte) (*http.Response, error) {
	ctx, span := c.tracer.Start(ctx, fmt.Sprintf("HTTP %s", method),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			semconv.HTTPURL(url),
		),
	)
	defer span.End()

	var resp *http.Response

	doRequest := func() error {
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
		if err != nil {
			return retry.Unrecoverable(err)
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		if c.debug {
			args := []any{"method", method, "url", url, "body", string(body), "headers", headers}
			if label, requestId := c.getRequestId(ctx); requestId != "" {
				args = append(args, label, requestId)
			}
			c.logger.Debug("http request", args...)

			if c.curlLog {
				command, _ := http2curl.GetCurlCommand(req)
				fmt.Println(command.String())
			}
		}

		var httpResp *http.Response
		if c.cb != nil {
			result, err := c.cb.Execute(func() (any, error) {
				r, err := c.handler.Do(req)
				if err != nil {
					return nil, err
				}
				if r.StatusCode >= 500 {
					return r, fmt.Errorf("server error: %d", r.StatusCode)
				}
				return r, nil
			})
			if err != nil {
				if errors.Is(err, gobreaker.ErrOpenState) {
					return retry.Unrecoverable(err)
				}
				return err
			}
			httpResp = result.(*http.Response)
		} else {
			httpResp, err = c.handler.Do(req)
			if err != nil {
				return err
			}
		}

		respBody, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return err
		}
		httpResp.Body.Close()
		httpResp.Body = io.NopCloser(bytes.NewReader(respBody))

		if c.debug {
			args := []any{"status", httpResp.StatusCode, "url", url, "body", string(respBody)}
			if label, requestId := c.getRequestId(ctx); requestId != "" {
				args = append(args, label, requestId)
			}
			c.logger.Debug("http response", args...)
		}

		resp = httpResp
		return nil
	}

	var err error
	if c.retry != nil {
		err = retry.Do(
			doRequest,
			retry.Attempts(c.retry.MaxAttempts),
			retry.Delay(c.retry.Delay),
			retry.MaxDelay(c.retry.MaxDelay),
			retry.DelayType(retry.BackOffDelay),
			retry.Context(ctx),
		)
	} else {
		err = doRequest()
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetAttributes(semconv.HTTPStatusCode(resp.StatusCode))
	if resp.StatusCode >= 400 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", resp.StatusCode))
	}

	return resp, nil
}

func (c *httpClient) getRequestId(ctx context.Context) (string, string) {
	if c.requestIdKey == "" {
		return "", ""
	}
	id, ok := ctx.Value(c.requestIdKey).(string)
	if !ok || id == "" {
		return "", ""
	}
	return c.requestIdLabel, id
}
