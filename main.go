package main

import (
	"errors"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/internal/updaters"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/parentError"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/routes"
)

func printError(err error) {
	var parentErr parentError.ParentError
	var contextedErr contextedError.ContextedError
	switch {
	case errors.As(err, &parentErr):
		fmt.Printf("%s: %s\r\n", parentErr.GetOrigin(), err.Error())
		printError(parentErr.GetChild())
	case errors.As(err, &contextedErr):
		fmt.Printf("%s: %s\r\n", contextedErr.GetOrigin(), err.Error())
	default:
		fmt.Println(err.Error())
	}
}

type App struct {
	config              map[string]interface{}
	domainRouteUpdater  updaters.DomainRouteUpdater
	domainRouteInterval int64
	keeneticUpdater     updaters.KeeneticUpdater
	keeneticIntercal    int64
}

func (a *App) Init() {
	// TODO: Database
	domains := []routes.DomainRoute{
		{Domain: "google.com"},
		{Domain: "yandex.ru"},
	}

	var ok bool
	var err error

	if ok, err = a.domainRouteUpdater.Init(a.config["domain.server"].(string), domains); err != nil {
		err = parentError.New("domainRouteUpdater initialization error", &err)
		printError(err)
		return
	}
	a.domainRouteUpdater.MaxTTL = a.config["domain.ttl.max"].(int64)
	a.domainRouteUpdater.MinTTL = a.config["domain.ttl.min"].(int64)
	a.domainRouteUpdater.DefaultTTL = a.config["domain.ttl.default"].(int64)
	a.SetDomainRouteInterval(a.config["domain.interval"].(int64))

	if ok, err = a.keeneticUpdater.Init(
		a.config["keenetic.host"].(string),
		a.config["keenetic.login"].(string),
		a.config["keenetic.password"].(string),
	); err != nil {
		err = parentError.New("keeneticUpdater initialization error", &err)
		printError(err)
		return
	}
	a.SetKeeneticInterval(a.config["keenetic.interval"].(int64))

	_, err = a.keeneticUpdater.Tick()
	if err != nil {
		printError(err)
	}

	fmt.Printf("%+v", a.keeneticUpdater.GetInterfaces())
	fmt.Printf("%+v", a.keeneticUpdater.GetRoutes())

	_ = ok
}

func (a *App) SetKeeneticInterval(sec int64) {
	a.keeneticIntercal = sec
}

func (a *App) SetDomainRouteInterval(sec int64) {
	a.domainRouteInterval = sec
}

func main() {
	app := App{
		config: map[string]interface{}{
			"domain.ttl.max":     int64(3600),
			"domain.ttl.min":     int64(60),
			"domain.ttl.default": int64(300),
			"domain.server":      "192.168.1.1",
			"domain.interval":    int64(5),
			"keenetic.host":      "https://keenetic.demo.keenetic.pro",
			"keenetic.login":     "demo",
			"keenetic.password":  "demo",
			"keenetic.interval":  int64(300),
		},
		domainRouteUpdater: *new(updaters.DomainRouteUpdater),
		keeneticUpdater:    *new(updaters.KeeneticUpdater),
	}
	app.Init()
}
