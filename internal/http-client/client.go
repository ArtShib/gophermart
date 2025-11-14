package http_client

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ArtShib/gophermart.git/internal/models"
)

type Client struct {
	log        *slog.Logger
	httpClient *http.Client
}

func New(log *slog.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Second * 10,
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
				MaxIdleConns:        100,
				MaxConnsPerHost:     10,
				IdleConnTimeout:     time.Second * 3,
			},
		},
		log: log,
	}
}

func (c *Client) RequestAccrualOrder(ctx context.Context, urlConnect string) (*models.ResAccrualOrder, error) {
	const op = "Client.RequestAccrualOrder"

	log := c.log.With(
		slog.String("op", op),
		slog.String("urlConnect", fmt.Sprintf("%v", urlConnect)),
	)

	log.Info("request urlConnect")

	nCtx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	request, err := http.NewRequestWithContext(nCtx, http.MethodGet, urlConnect, nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: invalid status code: %d", op, resp.StatusCode)
	}
	resAccrualOrder := models.ResAccrualOrder{}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&resAccrualOrder); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &resAccrualOrder, nil
}
