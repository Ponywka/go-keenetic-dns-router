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
	body, err := k.RCI([]interface{}{
		map[string]interface{}{
			"show": map[string]interface{}{
				"interface": map[string]interface{}{},
			},
		},
	})
	log.Printf("%+v", body)
	log.Println("EndTick")
	return true
}
