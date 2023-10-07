package updaters

import (
	dns "github.com/Focinfi/go-dns-resolver"
	"github.com/Ponywka/go-keenetic-dns-router/routes"
	"log"
	"reflect"
	"sort"
	"time"
)

type DomainRouteUpdater struct {
	MinTTL   int64
	MaxTTL   int64
	Resolver *dns.Resolver
	Domains  []routes.DomainRouteExtended
}

func (u *DomainRouteUpdater) Add(domainRoute routes.DomainRoute) {
	u.Domains = append(u.Domains, routes.DomainRouteExtended{DomainRoute: domainRoute})
}

func (u *DomainRouteUpdater) resolveDomain(domain *routes.DomainRouteExtended) bool {
	// Resolve domain
	resolver := u.Resolver
	resolver.Targets(domain.Domain).Types(dns.TypeA)
	result := resolver.Lookup()

	// Compute domain info
	UpdateTime := time.Now().Unix()
	LowestTTL := u.MaxTTL
	var IPs []string
	for target := range result.ResMap {
		for _, r := range result.ResMap[target] {
			TTLMillis := int64(r.Ttl.Seconds())
			if TTLMillis < LowestTTL {
				LowestTTL = TTLMillis
			}
			IPs = append(IPs, r.Content)
		}
	}
	if LowestTTL < u.MinTTL {
		LowestTTL = u.MinTTL
	}
	sort.Strings(IPs)

	// Updating domain info
	domain.LastResolved = UpdateTime
	domain.NextResolve = UpdateTime + LowestTTL
	if reflect.DeepEqual(IPs, domain.IPs) {
		return false
	}
	domain.IPs = IPs
	return true
}

func (u *DomainRouteUpdater) Tick() (bool, error) {
	log.Println("Tick")
	for index := range u.Domains {
		domain := &u.Domains[index]
		if domain.NextResolve > time.Now().Unix() {
			continue
		}
		log.Printf("Processing %s domain", domain.Domain)
		ok := u.resolveDomain(domain)
		if ok {
			log.Printf("%s IPs: %+v", domain.Domain, domain.IPs)
		}
	}
	log.Println("EndTick")
	return true, nil
}
