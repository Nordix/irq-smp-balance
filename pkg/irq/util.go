package irq

import (
	"os"
	"os/signal"
	"syscall"
)

// NewOSSignalChannel creates new os signal channel
func NewOSSignalChannel() chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		os.Interrupt,
		// More Linux signals here
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	return c
}
