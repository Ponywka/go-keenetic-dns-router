package main

import (
	"errors"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/errors/parentError"
	"github.com/Ponywka/go-keenetic-dns-router/routes"
	"github.com/Ponywka/go-keenetic-dns-router/updaters"
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
	config             map[string]string
	domainRouteUpdater updaters.DomainRouteUpdater
	keeneticUpdater    updaters.KeeneticUpdater
}

func (a *App) Init() {
	// TODO: Database
	domains := []routes.DomainRoute{
		{Domain: "google.com"},
		{Domain: "yandex.ru"},
	}

	var ok bool
	var err error

	if ok, err = a.domainRouteUpdater.Init(a.config["domain.server"], domains); err != nil {
		err = parentError.New("domainRouteUpdater initialization error", &err)
		printError(err)
		return
	}

	if ok, err = a.keeneticUpdater.Init(
		a.config["keenetic.host"],
		a.config["keenetic.login"],
		a.config["keenetic.password"],
	); err != nil {
		err = parentError.New("keeneticUpdater initialization error", &err)
		printError(err)
		return
	}

	_ = ok
}

func main() {
	app := App{
		config: map[string]string{
			"domain.server":     "192.168.1.1",
			"keenetic.host":     "https://keenetic.demo.keenetic.pro",
			"keenetic.login":    "demo",
			"keenetic.password": "demo",
		},
		domainRouteUpdater: *new(updaters.DomainRouteUpdater),
		keeneticUpdater:    *new(updaters.KeeneticUpdater),
	}
	app.Init()
}
