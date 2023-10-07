package main

import (
	"errors"
	"fmt"
	dns "github.com/Focinfi/go-dns-resolver"
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

func main() {
	// DNS
	domainRouteUpdater := updaters.DomainRouteUpdater{
		Resolver: dns.NewResolver("192.168.1.1"),
		MinTTL:   60,
		MaxTTL:   300,
	}
	domains := []string{"google.com", "yandex.ru"}
	for _, domain := range domains {
		domainRoute := routes.DomainRoute{Domain: domain, Comment: fmt.Sprintf("This domain is %s", domain)}
		domainRouteUpdater.Add(domainRoute)
	}
	_, err := domainRouteUpdater.Tick()
	if err != nil {
		printError(err)
	}

	// Keenetic
	keeneticUpdater := updaters.KeeneticUpdater{
		Login:    "demo",
		Password: "demo",
		URL:      "https://keenetic.demo.keenetic.pro",
	}
	_, err = keeneticUpdater.Tick()
	if err != nil {
		printError(err)
	}
}
