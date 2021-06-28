// Copyright (c) 2020-2021 Nordix Foundation.
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
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

const (
	irqSmpAffinityProcFile = "/tmp/default_smp_affinity"
	podIrqBannedCPUsFile   = "/tmp/pod_irq_banned_cpus"
)

func TestSetIRQLoadBalancing(t *testing.T) {
	g := NewGomegaWithT(t)

	fa, err := os.OpenFile(irqSmpAffinityProcFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = fa.Write([]byte("00ffffff,ffffffff"))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(fa.Close()).NotTo(HaveOccurred())
	defer func() {
		if err := os.Remove(irqSmpAffinityProcFile); err != nil {
			t.Errorf("error closing the file %s: %v", irqSmpAffinityProcFile, err)
		}
	}()

	fi, err := os.OpenFile(podIrqBannedCPUsFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(fi.Close()).NotTo(HaveOccurred())
	defer func() {
		if err := os.Remove(podIrqBannedCPUsFile); err != nil {
			t.Errorf("error closing the file %s: %v", podIrqBannedCPUsFile, err)
		}
	}()

	err = SetIRQLoadBalancing("1-2", false, irqSmpAffinityProcFile, podIrqBannedCPUsFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", false, irqSmpAffinityProcFile, podIrqBannedCPUsFile)
	g.Expect(err).NotTo(HaveOccurred())

	fa, err = os.OpenFile(irqSmpAffinityProcFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err := ioutil.ReadAll(fa)
	g.Expect(fa.Close()).NotTo(HaveOccurred())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("00ffffff,ffffffe0"))

	fb, err := os.OpenFile(podIrqBannedCPUsFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err = ioutil.ReadAll(fb)
	g.Expect(fb.Close()).NotTo(HaveOccurred())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("ff000000,0000001f"))
}

func TestResetIRQLoadBalancing(t *testing.T) {
	g := NewGomegaWithT(t)

	fa, err := os.OpenFile(irqSmpAffinityProcFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = fa.Write([]byte("00ffffff,ffffffff"))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(fa.Close()).NotTo(HaveOccurred())
	defer func() {
		if err := os.Remove(irqSmpAffinityProcFile); err != nil {
			t.Errorf("error closing the file %s: %v", irqSmpAffinityProcFile, err)
		}
	}()

	fi, err := os.OpenFile(podIrqBannedCPUsFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(fi.Close()).NotTo(HaveOccurred())
	defer func() {
		if err := os.Remove(podIrqBannedCPUsFile); err != nil {
			t.Errorf("error closing the file %s: %v", podIrqBannedCPUsFile, err)
		}
	}()

	err = SetIRQLoadBalancing("1-2", false, irqSmpAffinityProcFile, podIrqBannedCPUsFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", false, irqSmpAffinityProcFile, podIrqBannedCPUsFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("1-2", true, irqSmpAffinityProcFile, podIrqBannedCPUsFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", true, irqSmpAffinityProcFile, podIrqBannedCPUsFile)
	g.Expect(err).NotTo(HaveOccurred())

	fa, err = os.OpenFile(irqSmpAffinityProcFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err := ioutil.ReadAll(fa)
	g.Expect(fa.Close()).NotTo(HaveOccurred())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("00ffffff,ffffffff"))

	fb, err := os.OpenFile(podIrqBannedCPUsFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err = ioutil.ReadAll(fb)
	g.Expect(fb.Close()).NotTo(HaveOccurred())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("ff000000,00000000"))
}
