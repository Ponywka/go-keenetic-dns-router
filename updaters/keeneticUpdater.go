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
	ok, err := k.Auth(u.Login, u.Password)
	if err != nil {
		log.Printf("%+v", err)
		return false
	}
	if !ok {
		return false
	}
	var list []map[string]interface{}
	k.ToRciQueryList(&list, "show.interface", []map[string]interface{}{
		{"name": "Dsl0"},
	})
	body, err := k.Rci(list)
	log.Printf("%+v", body)
	log.Println("EndTick")
	return true
}
