package httpx

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var ErrBodyNotReplayable = errors.New("request body cannot be replayed for retry")

// RequestValidator rejects a request before it is sent.
type RequestValidator func(*http.Request) error

// RetryPolicy controls retries. MaxAttempts includes the initial request.
// Empty Methods and StatusCodes use conservative defaults.
type RetryPolicy struct {
	MaxAttempts          int
	BaseDelay            time.Duration
	MaxDelay             time.Duration
	Methods              []string
	StatusCodes          []int
	RetryTransportErrors bool
	RespectRetryAfter    bool
	OnRetry              func(RetryEvent)
}

type RetryEvent struct {
	Attempt     int
	NextAttempt int
	Method      string
	URL         string
	StatusCode  int
	Err         error
	Delay       time.Duration
}

func (p RetryPolicy) clone() RetryPolicy {
	p.Methods = append([]string(nil), p.Methods...)
	p.StatusCodes = append([]int(nil), p.StatusCodes...)
	return p
}

func (p RetryPolicy) normalized() RetryPolicy {
	p = p.clone()
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}
	if p.BaseDelay <= 0 {
		p.BaseDelay = 100 * time.Millisecond
	}
	if p.MaxDelay <= 0 {
		p.MaxDelay = 2 * time.Second
	}
	if len(p.Methods) == 0 {
		p.Methods = []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPut,
			http.MethodDelete,
		}
	}
	if len(p.StatusCodes) == 0 {
		p.StatusCodes = []int{
			http.StatusRequestTimeout,
			http.StatusTooEarly,
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		}
	}
	return p
}

func (p RetryPolicy) allowsMethod(method string) bool {
	for _, allowed := range p.Methods {
		if strings.EqualFold(method, allowed) {
			return true
		}
	}
	return false
}

func (p RetryPolicy) allowsStatus(statusCode int) bool {
	for _, allowed := range p.StatusCodes {
		if statusCode == allowed {
			return true
		}
	}
	return false
}

func (p RetryPolicy) delay(attempt int, response *http.Response) time.Duration {
	if p.RespectRetryAfter && response != nil {
		if delay, ok := parseRetryAfter(response.Header.Get("Retry-After"), time.Now()); ok {
			if delay > p.MaxDelay {
				return p.MaxDelay
			}
			return delay
		}
	}
	delay := p.BaseDelay
	for current := 1; current < attempt && delay < p.MaxDelay; current++ {
		if delay > p.MaxDelay/2 {
			return p.MaxDelay
		}
		delay *= 2
	}
	if delay > p.MaxDelay {
		return p.MaxDelay
	}
	return delay
}

func parseRetryAfter(value string, now time.Time) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second, true
	}
	when, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}
	delay := when.Sub(now)
	if delay < 0 {
		delay = 0
	}
	return delay, true
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
