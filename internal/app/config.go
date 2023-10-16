package app

type Config struct {
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
