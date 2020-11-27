package irq

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"unicode"

	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

const (
	// IrqSmpAffinityProcFile file containing irq mask settings
	IrqSmpAffinityProcFile = "/host/proc/irq/default_smp_affinity"
	// IrqBalanceConfigFile file containing irq balance banned cpus parameter
	IrqBalanceConfigFile = "/host/etc/sysconfig/podirqbalance"
	// IrqBalanceBannedCpus key for IRQBALANCE_BANNED_CPUS parameter
	IrqBalanceBannedCpus = "IRQBALANCE_BANNED_CPUS"
)

var mu sync.Mutex

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

// SetIRQLoadBalancing enable or disable the irq loadbalance on given cpus
func SetIRQLoadBalancing(cpus string, enable bool, irqSmpAffinityFile, irqBalanceConfigFile string) error {
	mu.Lock()
	defer mu.Unlock()

	content, err := ioutil.ReadFile(irqSmpAffinityFile)
	if err != nil {
		return err
	}
	currentIRQSMPSetting := strings.TrimSpace(string(content))
	newIRQSMPSetting, newIRQBalanceSetting, err := UpdateIRQSmpAffinityMask(cpus, currentIRQSMPSetting, enable)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(irqSmpAffinityFile, []byte(newIRQSMPSetting), 0o644); err != nil {
		return err
	}

	logrus.Infof("irqbalance banned cpus %s", newIRQBalanceSetting)

	irqBalanceConfig, err := os.Create(irqBalanceConfigFile)
	if err != nil {
		return err
	}
	defer irqBalanceConfig.Close()

	_, err = irqBalanceConfig.WriteString(newIRQBalanceSetting)
	if err != nil {
		return err
	}

	return nil
}

// The folloing methods are copied from github.com/cri-o/cri-o/internal/runtimehandlerhooks
// (reuse not possible as runtimehandlerhooks in internal package)

// UpdateIRQSmpAffinityMask take input cpus that need to change irq affinity mask and
// the current mask string, return an update mask string and inverted mask, with those cpus
// enabled or disable in the mask.
func UpdateIRQSmpAffinityMask(cpus, current string, set bool) (cpuMask, bannedCPUMask string, err error) {
	podcpuset, err := cpuset.Parse(cpus)
	if err != nil {
		return cpus, "", err
	}

	// only ascii string supported
	if !isASCII(current) {
		return cpus, "", fmt.Errorf("non ascii character detected: %s", current)
	}

	// remove ","; now each element is "0-9,a-f"
	s := strings.ReplaceAll(current, ",", "")

	// the index 0 corresponds to the cpu 0-7
	currentMaskArray, err := mapHexCharToByte(s)
	if err != nil {
		return cpus, "", err
	}
	invertedMaskArray := invertByteArray(currentMaskArray)

	for _, cpu := range podcpuset.ToSlice() {
		if set {
			// each byte represent 8 cpus
			currentMaskArray[cpu/8] |= cpuMaskByte(cpu % 8)
			invertedMaskArray[cpu/8] &^= cpuMaskByte(cpu % 8)
		} else {
			currentMaskArray[cpu/8] &^= cpuMaskByte(cpu % 8)
			invertedMaskArray[cpu/8] |= cpuMaskByte(cpu % 8)
		}
	}

	maskString := mapByteToHexChar(currentMaskArray)
	invertedMaskString := mapByteToHexChar(invertedMaskArray)

	maskStringWithComma := maskString[0:8]
	invertedMaskStringWithComma := invertedMaskString[0:8]
	for i := 8; i+8 <= len(maskString); i += 8 {
		maskStringWithComma = maskStringWithComma + "," + maskString[i:i+8]
		invertedMaskStringWithComma = invertedMaskStringWithComma + "," + invertedMaskString[i:i+8]
	}
	return maskStringWithComma, invertedMaskStringWithComma, nil
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func cpuMaskByte(c int) byte {
	return byte(1 << c)
}

func mapHexCharToByte(h string) ([]byte, error) {
	l := len(h)
	var hexin string
	if l%2 != 0 {
		// expect even number of chars
		hexin = "0" + h
	} else {
		hexin = h
	}

	breversed, err := hex.DecodeString(hexin)
	if err != nil {
		return nil, err
	}

	l = len(breversed)
	var barray []byte
	var rindex int
	for i := 0; i < l; i++ {
		rindex = l - i - 1
		barray = append(barray, breversed[rindex])
	}
	return barray, nil
}

func mapByteToHexChar(b []byte) string {
	var breversed []byte
	var rindex int
	l := len(b)
	// align it to 8 byte
	if l%8 != 0 {
		lfill := 8 - l%8
		l += lfill
		for i := 0; i < lfill; i++ {
			b = append(b, byte(0))
		}
	}

	for i := 0; i < l; i++ {
		rindex = l - i - 1
		breversed = append(breversed, b[rindex])
	}
	return hex.EncodeToString(breversed)
}

// take a byte array and invert each byte
func invertByteArray(in []byte) (out []byte) {
	for _, b := range in {
		out = append(out, byte(0xff)-b)
	}
	return
}
