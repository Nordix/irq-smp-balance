package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/pperiyasamy/irq-smp-balance/pkg/irq"
	"github.com/sirupsen/logrus"
)

// IrqBalanceConfigFile file containing irq balance banned cpus parameter
const (
	defaultIrqBalanceConfigFile = "/etc/sysconfig/podirqbalance"
	defaultLogFile              = "/var/log/irqsmpdaemon.log"
)

func main() {

	irqBalanceConfigFile := flag.String("config", defaultIrqBalanceConfigFile, "irq balance config file")
	logFile := flag.String("log", defaultLogFile, "log file")
	flag.Parse()

	c := irq.NewOSSignalChannel()

	if err := initializeLog(*logFile); err != nil {
		panic(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Fatal(err)
	}
	defer watcher.Close()

	logrus.Infof("using config file %s", *irqBalanceConfigFile)

	done := make(chan bool)
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

	if err = initializeConfigFile(*irqBalanceConfigFile); err != nil {
		logrus.Fatal(err)
	}
	if err = watcher.Add(*irqBalanceConfigFile); err != nil {
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
	logrus.Infof("irqbalance service is restarted")

	return nil
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
	}
	return err
}
