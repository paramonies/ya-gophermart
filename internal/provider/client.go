package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/paramonies/ya-gophermart/internal/store/dto"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

type AccrualClient struct {
	URL    string
	Client *http.Client
}

func NewAccrualClient(address string) *AccrualClient {
	return &AccrualClient{
		URL:    fmt.Sprintf("%s/api/orders", address),
		Client: &http.Client{},
	}
}

func (pc *AccrualClient) GetOrder(orderNumber int) (*dto.ProviderOrder, error) {
	url := fmt.Sprintf("%s/%d", pc.URL, orderNumber)
	log.Debug(context.Background(), "get order handler for provider", "address", url, "method", "GET")
	var order dto.ProviderOrder

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	bodyBin, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debug(context.Background(), "data from provider", "order", string(bodyBin))

	err = json.Unmarshal(bodyBin, &order)
	if err != nil {
		return nil, err
	}

	log.Debug(context.Background(), "order from provider", "order", string(bodyBin))
	return &order, nil
}
