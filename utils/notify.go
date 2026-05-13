package utils

import (
	"os"
	"os/signal"
)

// Notify ...
func Notify(c chan os.Signal, sig ...os.Signal) <-chan os.Signal {
	if c == nil {
		c = make(chan os.Signal, 1)
	}
	signal.Notify(c, sig...)
	return c
}
