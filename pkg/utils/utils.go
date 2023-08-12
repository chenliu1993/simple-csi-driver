package utils

import (
	"os"
)

func GenMultiChs(drivers []string, stopCh chan os.Signal) map[string]chan os.Signal {
	stopChs := make(map[string]chan os.Signal)
	for _, driver := range drivers {
		stopChs[driver] = make(chan os.Signal, 1)
	}
	return stopChs
}

func SendToMultiChs(stopChs map[string]chan os.Signal, stopCh chan os.Signal) {
	cacheSignal := <-stopCh
	for _, ch := range stopChs {
		ch <- cacheSignal
	}
}
