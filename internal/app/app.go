package app

import (
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/internal/routes"
	"github.com/Ponywka/go-keenetic-dns-router/internal/updaters"
	"time"
)

type App struct {
	domainRouteUpdater       updaters.DomainRouteUpdater
	domainRouteUpdaterTicker updaterTicker
	keeneticUpdater          updaters.KeeneticUpdater
	keeneticUpdaterTicker    updaterTicker
}

func New(config *Config) error {
	a := App{
		domainRouteUpdater: *new(updaters.DomainRouteUpdater),
		keeneticUpdater:    *new(updaters.KeeneticUpdater),
	}

	// TODO: Database
	domains := []routes.DomainRoute{
		{Domain: "google.com"},
		{Domain: "yandex.ru"},
	}

	var ok bool
	var err error

	if ok, err = a.domainRouteUpdater.Init(config.DomainServer, domains); err != nil {
		return fmt.Errorf("domainRouteUpdater initialization error: %w", err)
	}
	a.domainRouteUpdater.MaxTTL = config.DomainTtlMax
	a.domainRouteUpdater.MinTTL = config.DomainTtlMin
	a.domainRouteUpdater.DefaultTTL = config.DomainTtlDefault
	domainRouteUpdaterTicker := createUpdaterTicker(&a.domainRouteUpdater, time.Duration(config.DomainInterval)*time.Second)
	go func() {
		for {
			res := <-domainRouteUpdaterTicker.Result
			if res.Error != nil {
				fmt.Println(res.Error.Error())
				continue
			}
			if res.Result {
				fmt.Println("Updated!")
			}
		}
	}()

	if ok, err = a.keeneticUpdater.Init(
		config.KeeneticHost,
		config.KeeneticLogin,
		config.KeeneticPassword,
	); err != nil {
		return fmt.Errorf("keeneticUpdater initialization error: %w", err)
	}
	keeneticUpdaterTicker := createUpdaterTicker(&a.keeneticUpdater, time.Duration(config.KeeneticInterval)*time.Second)
	go func() {
		for {
			res := <-keeneticUpdaterTicker.Result
			if res.Error != nil {
				fmt.Println(res.Error.Error())
				continue
			}
			if res.Result {
				fmt.Println("Updated!")
			}
		}
	}()

	time.Sleep(8 * time.Second)
	domainRouteUpdaterTicker.TickerReset <- 1 * time.Second
	domainRouteUpdaterTicker.TickerReset <- 1 * time.Second
	time.Sleep(5 * time.Second)
	domainRouteUpdaterTicker.TickerStop <- true
	domainRouteUpdaterTicker.TickerStop <- true
	time.Sleep(1 * time.Second)

	//_, err = a.keeneticUpdater.Tick()
	//if err != nil {
	//	fmt.Println(err.Error())
	//}

	//fmt.Printf("%+v", a.keeneticUpdater.GetInterfaces())
	//fmt.Printf("%+v", a.keeneticUpdater.GetRoutes())

	_ = ok

	return nil
}

func (a *App) SetKeeneticInterval(sec int64) {
	a.domainRouteUpdaterTicker.TickerReset <- time.Duration(sec) * time.Second
}

func (a *App) SetDomainRouteInterval(sec int64) {
	a.domainRouteUpdaterTicker.TickerReset <- time.Duration(sec) * time.Second
}
