package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/mmeow0/gophermart-bonus/internal/model"
)

var (
	ErrTooManyRequests = errors.New("too many requests")
	ErrNotRegistered   = errors.New("order not registered")
	ErrAccrualService  = errors.New("accrual service error")
)

type AccrualClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAccrualClient(baseURL string) *AccrualClient {
	return &AccrualClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *AccrualClient) GetOrderAccrual(ctx context.Context, orderNumber string) (*model.AccrualResponse, int, error) {
	url := fmt.Sprintf("%s/api/orders/%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrualResp model.AccrualResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, err
		}

		if err := json.Unmarshal(body, &accrualResp); err != nil {
			return nil, resp.StatusCode, err
		}

		return &accrualResp, resp.StatusCode, nil

	case http.StatusNoContent:
		return nil, resp.StatusCode, ErrNotRegistered

	case http.StatusTooManyRequests:
		retryAfter := 60
		if retryHeader := resp.Header.Get("Retry-After"); retryHeader != "" {
			if seconds, err := strconv.Atoi(retryHeader); err == nil {
				retryAfter = seconds
			}
		}
		return nil, retryAfter, ErrTooManyRequests

	case http.StatusInternalServerError:
		return nil, resp.StatusCode, ErrAccrualService

	default:
		return nil, resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
