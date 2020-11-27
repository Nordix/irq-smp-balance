package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pperiyasamy/irq-smp-balance/pkg/irq"
	"github.com/sirupsen/logrus"
)

func main() {
	c := irq.NewOSSignalChannel()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					content, err := ioutil.ReadFile(irq.IrqSmpAffinityProcFile)
					if err != nil {
						logrus.Infof("error reading %s file : %v", irq.IrqSmpAffinityProcFile, err)
						return
					}
					if err = resetIRQBalance(strings.TrimSpace(string(content))); err != nil {
						logrus.Infof("irqbalance with banned cpus failed: %v", err)
					}

				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Infof("file watch error occurred: %v", err)
			}
		}
	}()

	err = watcher.Add(irq.IrqBalanceConfigFile)
	if err != nil {
		logrus.Fatal(err)
	}

	// Capture signals to cleanup before exiting
	<-c
	<-done
	logrus.Infof("irq smp daemon is stopped")
}

func resetIRQBalance(newIRQBalanceSetting string) error {
	logrus.Infof("restart irqbalance with banned cpus %s", newIRQBalanceSetting)
	cmd1 := exec.Command("service", "irqbalance", "restart")
	additionalEnv := "IRQBALANCE_BANNED_CPUS=" + newIRQBalanceSetting
	cmd1.Env = append(os.Environ(), additionalEnv)
	if err := cmd1.Run(); err != nil {
		logrus.Errorf("error restarting irqbalance service: error %v", err)
		cmd2 := exec.Command("irqbalance", "--oneshot")
		cmd2.Env = cmd1.Env
		return cmd2.Run()
	}
	return nil
}
