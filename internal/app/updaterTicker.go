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
	TickerReset chan time.Duration
	TickerStop  chan bool
	Result      chan updaterTickResult
}

func createUpdaterTicker(updater interfaces.Updater, d time.Duration) (ut updaterTicker) {
	ticker := time.NewTicker(d)
	ut = updaterTicker{
		TickerReset: make(chan time.Duration),
		TickerStop:  make(chan bool),
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
			case d := <-ut.TickerReset:
				ticker.Reset(d)
			case <-ut.TickerStop:
				ticker.Stop()
			}
		}
	}()
	return
}
