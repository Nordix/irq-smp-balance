// Copyright 2020 Ericsson Software Technology.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package irq

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"unicode"

	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

const (
	// IrqSmpAffinityProcFile file containing irq mask settings
	IrqSmpAffinityProcFile = "/host/proc/irq/default_smp_affinity"
	// PodIrqBannedCPUsFile file containing irq balance banned cpus parameter
	PodIrqBannedCPUsFile = "/host/etc/sysconfig/pod_irq_banned_cpus"
	// IrqBalanceBannedCpus key for IRQBALANCE_BANNED_CPUS parameter
	IrqBalanceBannedCpus = "IRQBALANCE_BANNED_CPUS"
)

var mu sync.Mutex

// SetIRQLoadBalancing enable or disable the irq loadbalance on given cpus
func SetIRQLoadBalancing(cpus string, enable bool, irqSmpAffinityFile, podIrqBannedCPUsFile string) error {
	mu.Lock()
	defer mu.Unlock()

	currentIRQSMPSetting, err := RetrieveCPUMask(irqSmpAffinityFile)
	if err != nil {
		return err
	}

	newIRQSMPSetting, newIRQBalanceSetting, err := UpdateIRQSmpAffinityMask(cpus, currentIRQSMPSetting, enable)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(irqSmpAffinityFile, []byte(newIRQSMPSetting), 0o644); err != nil {
		return err
	}

	logrus.Infof("irqbalance banned cpus %s", newIRQBalanceSetting)

	// write to pod cpu banned file at last so that fnotify write event triggered at right time.
	podIrqBannedCPUsConfig, err := os.Create(podIrqBannedCPUsFile)
	if err != nil {
		return err
	}
	defer podIrqBannedCPUsConfig.Close()

	_, err = podIrqBannedCPUsConfig.WriteString(newIRQBalanceSetting)
	if err != nil {
		return err
	}

	return nil
}

func updateIrqBalanceConfigFile(irqBalanceConfigFile, newIRQBalanceSetting string) error {
	input, err := ioutil.ReadFile(irqBalanceConfigFile)
	if err != nil {
		logrus.Infof("irqbalance config file %s doesn't exist", irqBalanceConfigFile)
		return nil
	}
	lines := strings.Split(string(input), "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, IrqBalanceBannedCpus+"=") {
			lines[i] = IrqBalanceBannedCpus + "=" + "\"" + newIRQBalanceSetting + "\""
			found = true
		}
	}
	output := strings.Join(lines, "\n")
	if !found {
		output = output + "\n" + IrqBalanceBannedCpus + "=" + "\"" + newIRQBalanceSetting + "\"" + "\n"
	}
	if err = ioutil.WriteFile(irqBalanceConfigFile, []byte(output), 0644); err != nil {
		return err
	}
	return nil
}

// ResetIRQBalance restart irqbalance daemon with newIRQBalanceSetting
func ResetIRQBalance(irqBalanceConfigFile, newIRQBalanceSetting string) error {
	logrus.Infof("restart irqbalance with banned cpus %s", newIRQBalanceSetting)
	if err := updateIrqBalanceConfigFile(irqBalanceConfigFile, newIRQBalanceSetting); err != nil {
		return err
	}
	cmd1 := exec.Command("service", "irqbalance", "restart")
	if err := cmd1.Run(); err != nil {
		logrus.Errorf("error restarting irqbalance service: error %v", err)
		cmd2 := exec.Command("irqbalance", "--oneshot")
		additionalEnv := IrqBalanceBannedCpus + "=" + newIRQBalanceSetting
		cmd2.Env = append(os.Environ(), additionalEnv)
		return cmd2.Run()
	}
	logrus.Infof("irqbalance service is restarted")

	return nil
}

// InvertMaskStringWithComma invert the give mask string retaining the comma
func InvertMaskStringWithComma(maskStringWithComma string) (string, error) {
	// only ascii string supported
	if !isASCII(maskStringWithComma) {
		return "", fmt.Errorf("non ascii character detected: %s", maskStringWithComma)
	}

	s := strings.ReplaceAll(maskStringWithComma, ",", "")
	currentMaskArray, err := mapHexCharToByte(s)
	if err != nil {
		return "", err
	}

	invertedMaskString := mapByteToHexChar(invertByteArray(currentMaskArray))
	invertedMaskStringWithComma := invertedMaskString[0:8]
	for i := 8; i+8 <= len(invertedMaskString); i += 8 {
		invertedMaskStringWithComma = invertedMaskStringWithComma + "," + invertedMaskString[i:i+8]
	}

	return invertedMaskStringWithComma, nil
}

// RetrieveCPUMask retrieves cpu masks set in irq smp affinity file
func RetrieveCPUMask(irqSmpAffinityFile string) (string, error) {
	content, err := ioutil.ReadFile(irqSmpAffinityFile)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
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
