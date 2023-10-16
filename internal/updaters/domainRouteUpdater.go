package updaters

import (
	dns "github.com/Focinfi/go-dns-resolver"
	"github.com/Ponywka/go-keenetic-dns-router/pkg/routes"
	"log"
	"math"
	"reflect"
	"sort"
	"time"
)

type DomainRouteUpdater struct {
	MinTTL     int64
	MaxTTL     int64
	DefaultTTL int64
	Resolver   *dns.Resolver
	Domains    map[string]routes.DomainRouteExtended
}

func (u *DomainRouteUpdater) Add(domainRoute routes.DomainRoute) {
	u.Domains[domainRoute.Domain] = routes.DomainRouteExtended{DomainRoute: domainRoute}
}

func (u *DomainRouteUpdater) resolveDomain(domain *routes.DomainRouteExtended) bool {
	// Resolve domain
	resolver := u.Resolver
	resolver.Targets(domain.Domain).Types(dns.TypeA)
	result := resolver.Lookup()

	// Compute domain info
	UpdateTime := time.Now().Unix()
	var IPs []string
	LowestTTL := u.DefaultTTL
	if len(result.ResMap) != 0 {
		LowestTTL = math.MaxInt64
		if u.MaxTTL != 0 {
			LowestTTL = u.MaxTTL
		}
		for target := range result.ResMap {
			for _, r := range result.ResMap[target] {
				TTLSeconds := int64(r.Ttl.Seconds())
				if TTLSeconds < LowestTTL {
					LowestTTL = TTLSeconds
				}
				IPs = append(IPs, r.Content)
			}
		}
		if u.MinTTL != 0 && LowestTTL < u.MinTTL {
			LowestTTL = u.MinTTL
		}
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

func (u *DomainRouteUpdater) Init(dnsServer string, domains []routes.DomainRoute) (bool, error) {
	u.Resolver = dns.NewResolver(dnsServer)
	u.Domains = make(map[string]routes.DomainRouteExtended)
	return true, nil
}

func (u *DomainRouteUpdater) Tick() (bool, error) {
	log.Println("Tick")
	isUpdated := false
	for _, domainRoute := range u.Domains {
		if domainRoute.NextResolve > time.Now().Unix() {
			continue
		}
		isUpdated = isUpdated || u.resolveDomain(&domainRoute)
	}
	if isUpdated {

	}
	log.Println("EndTick")
	return true, nil
}
