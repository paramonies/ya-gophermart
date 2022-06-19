package job

import (
	"context"
	"fmt"
	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/pkg/log"
	"time"
)

const jobRunInterval = 3 * time.Second

type Job struct {
	client  *provider.AccrualClient
	storage store.Connector
	done    chan struct{}
}

func InitJob(cli *provider.AccrualClient, sto store.Connector, done chan struct{}) *Job {
	return &Job{
		client:  cli,
		storage: sto,
		done:    done,
	}
}

func (j *Job) Run() {
	ticker := time.NewTicker(jobRunInterval)
	go func() {
		for {
			select {
			case <-j.done:
				return
			case <-ticker.C:
				j.loadAccruals()
			}
		}
	}()
}

func (j *Job) loadAccruals() {
	log.Info(context.Background(), "try to get accruals for orders")

	list, err := j.storage.Accruals().GetPendingOrders()
	if err != nil {
		log.Info(context.Background(), "failed to get pending orders")
	}

	if len(*list) != 0 && err == nil {
		go func() {
			for _, or := range *list {
				err := j.client.UpdateAccrual(or.OrderNumber)
				if err != nil {
					log.Error(context.Background(), fmt.Sprintf("failed to update %s order", or.ID), err)
				}
			}
		}()
	}
}
