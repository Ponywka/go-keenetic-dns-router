package routes

type DomainRouteExtended struct {
	DomainRoute
	IsResolved   bool
	LastResolved int64
	NextResolve  int64
	IPs          []string
}
