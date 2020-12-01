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
	defaultIrqBalanceConfigFile = "/etc/sysconfig/pod_irq_banned_cpus"
	defaultLogFile              = "/var/log/irqsmpdaemon.log"
)

func main() {

	irqBalanceConfigFile := flag.String("config", defaultIrqBalanceConfigFile, "pod irq banned cpus file")
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

	logrus.Infof("using config file %s", *irqBalanceConfigFile)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					content, err := ioutil.ReadFile(*irqBalanceConfigFile)
					if err != nil {
						logrus.Infof("error reading %s file : %v", *irqBalanceConfigFile, err)
						return
					}
					if err = irq.ResetIRQBalance(strings.TrimSpace(string(content))); err != nil {
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

	if err = initializeConfigFile(*irqBalanceConfigFile); err != nil {
		logrus.Fatal(err)
	}
	if err = watcher.Add(*irqBalanceConfigFile); err != nil {
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

func initializeConfigFile(configFile string) error {
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		irqBalanceConfig, err := os.Create(configFile)
		if err != nil {
			return err
		}
		irqBalanceConfig.Close()
		return nil
	} else if err == nil {
		content, err := ioutil.ReadFile(configFile)
		if err != nil {
			logrus.Infof("error reading %s file : %v", configFile, err)
			return err
		}
		if err = irq.ResetIRQBalance(strings.TrimSpace(string(content))); err != nil {
			logrus.Infof("irqbalance with banned cpus failed: %v", err)
		}
	}
	return err
}
