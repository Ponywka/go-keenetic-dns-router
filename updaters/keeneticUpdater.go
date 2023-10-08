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
	body, err := k.GetInterfaceList()
	if err != nil {
		return false, err
	}
	log.Printf("%+v", body)
	log.Println("EndTick")
	return true, nil
}
