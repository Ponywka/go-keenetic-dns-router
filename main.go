package main

import (
	"errors"
	"fmt"
	"github.com/Ponywka/go-keenetic-dns-router/internal/interfaces"
	"github.com/Ponywka/go-keenetic-dns-router/internal/updaters"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/errors/parentError"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/routes"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"time"
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

type AppConfig struct {
	DomainTtlMax     int64  `env:"DOMAIN_TTL_MAX,required"`
	DomainTtlMin     int64  `env:"DOMAIN_TTL_MIN,required"`
	DomainTtlDefault int64  `env:"DOMAIN_TTL_DEFAULT,required"`
	DomainServer     string `env:"DOMAIN_SERVER,required"`
	DomainInterval   int64  `env:"DOMAIN_INTERVAL,required"`
	KeeneticHost     string `env:"KEENETIC_HOST,required"`
	KeeneticLogin    string `env:"KEENETIC_LOGIN,required"`
	KeeneticPassword string `env:"KEENETIC_PASSWORD,required"`
	KeeneticInterval int64  `env:"KEENETIC_INTERVAL,required"`
}

type App struct {
	domainRouteUpdater  updaters.DomainRouteUpdater
	domainRouteInterval time.Duration
	keeneticUpdater     updaters.KeeneticUpdater
	keeneticIntercal    time.Duration
}

func (a *App) createUpdaterTicker(updater interfaces.Updater, d time.Duration) (timerUpdate chan time.Duration, quit chan bool) {
	ticker := time.NewTicker(d)
	timerUpdate = make(chan time.Duration)
	quit = make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				// TODO: Error handler
				_, _ = updater.Tick()
			case d := <-timerUpdate:
				ticker.Reset(d)
			case <-quit:
				return
			}
		}
	}()
	return
}

func (a *App) Init(config AppConfig) {
	// TODO: Database
	domains := []routes.DomainRoute{
		{Domain: "google.com"},
		{Domain: "yandex.ru"},
	}

	var ok bool
	var err error

	if ok, err = a.domainRouteUpdater.Init(config.DomainServer, domains); err != nil {
		err = parentError.New("domainRouteUpdater initialization error", &err)
		printError(err)
		return
	}
	a.domainRouteUpdater.MaxTTL = config.DomainTtlMax
	a.domainRouteUpdater.MinTTL = config.DomainTtlMin
	a.domainRouteUpdater.DefaultTTL = config.DomainTtlDefault
	a.SetDomainRouteInterval(time.Duration(config.DomainInterval))
	domainRouteUpdaterTimer, domainRouteUpdaterQuit := a.createUpdaterTicker(&a.domainRouteUpdater, a.domainRouteInterval*time.Second)

	if ok, err = a.keeneticUpdater.Init(
		config.KeeneticHost,
		config.KeeneticLogin,
		config.KeeneticPassword,
	); err != nil {
		err = parentError.New("keeneticUpdater initialization error", &err)
		printError(err)
		return
	}
	a.SetKeeneticInterval(time.Duration(config.KeeneticInterval))
	keeneticUpdaterTimer, keeneticUpdaterQuit := a.createUpdaterTicker(&a.keeneticUpdater, a.keeneticIntercal*time.Second)

	time.Sleep(8 * time.Second)
	domainRouteUpdaterTimer <- 1 * time.Second
	keeneticUpdaterTimer <- 1 * time.Second
	time.Sleep(5 * time.Second)
	domainRouteUpdaterQuit <- true
	keeneticUpdaterQuit <- true
	time.Sleep(1 * time.Second)

	//_, err = a.keeneticUpdater.Tick()
	//if err != nil {
	//	printError(err)
	//}

	//fmt.Printf("%+v", a.keeneticUpdater.GetInterfaces())
	//fmt.Printf("%+v", a.keeneticUpdater.GetRoutes())

	_ = ok
}

func (a *App) SetKeeneticInterval(sec time.Duration) {
	a.keeneticIntercal = sec
}

func (a *App) SetDomainRouteInterval(sec time.Duration) {
	a.domainRouteInterval = sec
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}

	config := AppConfig{}
	if err := env.Parse(&config); err != nil {
		panic(err)
	}

	app := App{
		domainRouteUpdater: *new(updaters.DomainRouteUpdater),
		keeneticUpdater:    *new(updaters.KeeneticUpdater),
	}
	app.Init(config)
}
