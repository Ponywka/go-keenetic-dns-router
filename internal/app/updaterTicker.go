package app

import (
	"github.com/Ponywka/go-keenetic-dns-router/internal/interfaces"
	"time"
)

type updaterTickResult struct {
	Result bool
	Error  error
}

type updaterTicker struct {
	TimerUpdate chan time.Duration
	Quit        chan bool
	Result      chan updaterTickResult
}

func createUpdaterTicker(updater interfaces.Updater, d time.Duration) (ut updaterTicker) {
	ticker := time.NewTicker(d)
	ut = updaterTicker{
		TimerUpdate: make(chan time.Duration),
		Quit:        make(chan bool),
		Result:      make(chan updaterTickResult),
	}
	go func() {
		for {
			select {
			case <-ticker.C:
				res, err := updater.Tick()
				ut.Result <- updaterTickResult{
					Result: res,
					Error:  err,
				}
			case d := <-ut.TimerUpdate:
				ticker.Reset(d)
			case <-ut.Quit:
				ticker.Stop()
				break
			}
		}
	}()
	return
}
