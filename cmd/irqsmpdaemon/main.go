package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/pperiyasamy/irq-smp-balance/pkg/irq"
	"github.com/sirupsen/logrus"
)

const (
	defaultPodIrqBannedCPUsFile = "/etc/sysconfig/pod_irq_banned_cpus"
	defaultIrqBalanceConfigFile = "/etc/sysconfig/irqbalance"
	defaultLogFile              = "/var/log/irqsmpdaemon.log"
)

func main() {

	podIrqBannedCPUsFile := flag.String("podfile", defaultPodIrqBannedCPUsFile, "pod irq banned cpus file")
	irqBalanceConfigFile := flag.String("config", defaultIrqBalanceConfigFile, "irq balance config file")
	logFile := flag.String("log", defaultLogFile, "log file")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	done := make(chan bool, 1)

	if err := initializeLog(*logFile); err != nil {
		panic(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal(err)
	}
	defer watcher.Close()

	logrus.Infof("using config file %s", *podIrqBannedCPUsFile)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					content, err := ioutil.ReadFile(*podIrqBannedCPUsFile)
					if err != nil {
						logrus.Infof("error reading %s file : %v", *podIrqBannedCPUsFile, err)
						return
					}
					if err = irq.ResetIRQBalance(*irqBalanceConfigFile, strings.TrimSpace(string(content))); err != nil {
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

	if err = initializeConfigFile(*podIrqBannedCPUsFile, *irqBalanceConfigFile); err != nil {
		logrus.Fatal(err)
	}
	if err = watcher.Add(*podIrqBannedCPUsFile); err != nil {
		logrus.Fatal(err)
	}

	go func() {
		sig := <-sigs
		logrus.Infof("received the signal %v", sig)
		done <- true
	}()

	// Capture signals to cleanup before exiting
	<-done

	logrus.Infof("irq smp daemon is stopped")
}

func initializeLog(logFile string) error {
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	logrus.SetOutput(f)
	return nil
}

func initializeConfigFile(podIrqBannedCPUsFile, irqBalanceConfigFile string) error {
	_, err := os.Stat(podIrqBannedCPUsFile)
	if os.IsNotExist(err) {
		irqBalanceConfig, err := os.Create(podIrqBannedCPUsFile)
		if err != nil {
			return err
		}
		irqBalanceConfig.Close()
		return nil
	} else if err == nil {
		content, err := ioutil.ReadFile(podIrqBannedCPUsFile)
		if err != nil {
			logrus.Infof("error reading %s file : %v", podIrqBannedCPUsFile, err)
			return err
		}
		if err = irq.ResetIRQBalance(irqBalanceConfigFile, strings.TrimSpace(string(content))); err != nil {
			logrus.Infof("irqbalance with banned cpus failed: %v", err)
		}
	}
	return err
}
