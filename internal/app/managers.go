package app

import (
	"github.com/paramonies/ya-gophermart/internal/managers"
	"github.com/paramonies/ya-gophermart/internal/store"
)

//type T struct{}
// Verify that T implements I.
//var _ I = T{}
// Verify that *T implements I.
//var _ I = (*T)(nil)

var _ managers.AppManagers = (*appManagers)(nil)

type appManagers struct {
	userManager    managers.UserManager
	orderManager   managers.OrderManager
	accrualManager managers.AccrualManager
}

func NewAppManagers(storage store.Connector) managers.AppManagers {
	userManager := NewUserManager(storage)
	orderManager := NewOrderManager(storage)
	accrualManager := NewAccrualManager(storage)
	return &appManagers{
		userManager:    userManager,
		orderManager:   orderManager,
		accrualManager: accrualManager,
	}
}

func (am *appManagers) UserManager() managers.UserManager {
	return am.userManager
}

func (am *appManagers) OrderManager() managers.OrderManager {
	return am.orderManager
}

func (am *appManagers) AccrualManager() managers.AccrualManager {
	return am.accrualManager
}
