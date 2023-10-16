package updaters

import (
	"errors"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/keenetic"
	"log"
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

func (u *KeeneticUpdater) Init(host, login, password string) (bool, error) {
	u.client = keenetic.New(host)
	ok, err := u.client.Auth(login, password)
	if err != nil {
		return false, fmt.Errorf("auth error: %w", err)
	}
	if !ok {
		return false, errors.New("login or password invalid")
	}
	u.interfaces, err = u.client.GetInterfaceList()
	if err != nil {
		return false, errors.New("getting interfaces error")
	}
	u.routes, err = u.client.GetRouteList()
	if err != nil {
		return false, errors.New("getting routes error")
	}
	return true, nil
}

func (u *KeeneticUpdater) Tick() (bool, error) {
	log.Println("Tick")
	log.Println("EndTick")
	return true, nil
}
