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
	ticker      *time.Ticker
}

func createUpdaterTicker(updater interfaces.Updater, d time.Duration) (ut updaterTicker) {
	ut = updaterTicker{
		TickerReset: make(chan time.Duration),
		TickerStop:  make(chan bool),
		Result:      make(chan updaterTickResult),
		ticker:      time.NewTicker(d),
	}
	go func() {
		for {
			select {
			case <-ut.ticker.C:
				res, err := updater.Tick()
				ut.Result <- updaterTickResult{
					Result: res,
					Error:  err,
				}
			case d := <-ut.TickerReset:
				ut.ticker.Reset(d)
			case <-ut.TickerStop:
				ut.ticker.Stop()
			}
		}
	}()
	return
}
