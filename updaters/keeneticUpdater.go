package updaters

import (
	"github.com/Ponywka/go-keenetic-dns-router/keenetic"
	"log"
)

type KeeneticUpdater struct {
	Login       string
	Password    string
	URL         string
	isConnected bool
}

func (u *KeeneticUpdater) Tick() bool {
	log.Println("Tick")
	k := keenetic.NewKeeneticClient(u.URL)
	res, err := k.Auth(u.Login, u.Password)
	if err != nil {
		log.Printf("%+v", err)
	}
	log.Printf("%+v", res)
	log.Println("EndTick")
	return true
}
