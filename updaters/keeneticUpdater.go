package updaters

import (
	"github.com/Ponywka/go-keenetic-dns-router/errors/contextedError"
	"github.com/Ponywka/go-keenetic-dns-router/keenetic"
	"log"
)

type KeeneticUpdater struct {
	Login       string
	Password    string
	URL         string
	isConnected bool
}

func (u *KeeneticUpdater) Tick() (bool, error) {
	log.Println("Tick")
	k := keenetic.NewKeeneticClient(u.URL)
	ok, err := k.Auth(u.Login, u.Password)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, contextedError.New("login or password invalid")
	}
	var list []map[string]interface{}
	err = k.ToRciQueryList(&list, "show.interface", []map[string]interface{}{
		{"name": "Dsl0"},
	})
	if err != nil {
		return false, err
	}
	body, err := k.Rci(list)
	log.Printf("%+v", body)
	log.Println("EndTick")
	return true, nil
}
