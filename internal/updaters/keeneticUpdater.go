package updaters

import (
	"errors"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/keenetic"
	"reflect"
)

type KeeneticUpdater struct {
	URL         string
	isConnected bool
	client      keenetic.Client
	interfaces  []keenetic.InterfaceBase
	routes      []keenetic.Route
}

func (u *KeeneticUpdater) GetInterfaces() []keenetic.InterfaceBase {
	return u.interfaces
}

func (u *KeeneticUpdater) GetRoutes() []keenetic.Route {
	return u.routes
}

func (u *KeeneticUpdater) Init(host, login, password string) error {
	u.client = keenetic.New(host)
	ok, err := u.client.Auth(login, password)
	if err != nil {
		return fmt.Errorf("auth error: %w", err)
	}
	if !ok {
		return errors.New("login or password invalid")
	}
	u.interfaces, err = u.client.GetInterfaceList()
	if err != nil {
		return fmt.Errorf("getting interfaces error: %w", err)
	}
	u.routes, err = u.client.GetRouteList()
	if err != nil {
		return fmt.Errorf("getting routes error: %w", err)
	}
	return nil
}

func (u *KeeneticUpdater) Tick() (bool, error) {
	interfaces, err := u.client.GetInterfaceList()
	if err != nil {
		return false, fmt.Errorf("getting interfaces error: %w", err)
	}
	routes, err := u.client.GetRouteList()
	if err != nil {
		return false, fmt.Errorf("getting routes error: %w", err)
	}

	isUpdated := false
	if !reflect.DeepEqual(interfaces, u.interfaces) {
		u.interfaces = interfaces
		isUpdated = true
	}
	if !reflect.DeepEqual(routes, u.routes) {
		u.routes = routes
		isUpdated = true
	}

	return isUpdated, nil
}
