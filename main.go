package main

import (
	"fmt"
	dns "github.com/Focinfi/go-dns-resolver"
	"github.com/Ponywka/go-keenetic-dns-router/routes"
	"github.com/Ponywka/go-keenetic-dns-router/updaters"
)

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
	domainRouteUpdater.Tick()

	// Keenetic
	keeneticUpdater := updaters.KeeneticUpdater{
		Login:    "demo",
		Password: "demo",
		URL:      "https://keenetic.demo.keenetic.pro",
	}
	keeneticUpdater.Tick()
}
