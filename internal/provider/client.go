package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/internal/store/dto"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

type AccrualClient struct {
	URL     string
	Client  *http.Client
	Storage store.Connector
}

func NewAccrualClient(address string, storage store.Connector) *AccrualClient {
	return &AccrualClient{
		URL:     fmt.Sprintf("%s/api/orders", address),
		Client:  &http.Client{},
		Storage: storage,
	}
}

func (ac *AccrualClient) getOrder(orderNumber string) (*dto.ProviderOrder, error) {
	url := fmt.Sprintf("%s/%s", ac.URL, orderNumber)
	log.Debug(context.Background(), "get order handler for provider", "address", url, "method", "GET")
	var order dto.ProviderOrder

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := ac.Client.Do(req)
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

func (ac *AccrualClient) UpdateAccrual(orderNumber string) error {
	order, err := ac.getOrder(orderNumber)
	if err != nil {
		return err
	}

	if order == nil {
		log.Debug(context.Background(), "no information for order in accrual server provider", "orderNumber", orderNumber)
		return nil
	}

	err = ac.Storage.Accruals().UpdateAccrual(*order)
	if err != nil {
		return err
	}

	log.Debug(context.Background(), "load accrual for order", "orderNumber", orderNumber, "Accrual", order.Accrual)
	return nil
}
