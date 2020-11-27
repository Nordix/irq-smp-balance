package irq

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestSetIRQLoadBalancing(t *testing.T) {
	g := NewGomegaWithT(t)

	irqSmpAffinityProcFile := "/tmp/default_smp_affinity"
	irqBalanceConfigFile := "/tmp/irqbalance"

	fa, err := os.OpenFile(irqSmpAffinityProcFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	fa.Write([]byte("00ffffff,ffffffff"))
	fa.Close()
	defer os.Remove(irqSmpAffinityProcFile)

	fi, err := os.OpenFile("/tmp/irqbalance", os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	fi.Close()
	defer os.Remove(irqBalanceConfigFile)

	err = SetIRQLoadBalancing("1-2", false, irqSmpAffinityProcFile, irqBalanceConfigFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", false, irqSmpAffinityProcFile, irqBalanceConfigFile)
	g.Expect(err).NotTo(HaveOccurred())

	fa, err = os.OpenFile(irqSmpAffinityProcFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err := ioutil.ReadAll(fa)
	fa.Close()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("00ffffff,ffffffe0"))

	fb, err := os.OpenFile(irqBalanceConfigFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err = ioutil.ReadAll(fb)
	fb.Close()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("ff000000,0000001f"))
}

func TestResetIRQLoadBalancing(t *testing.T) {
	g := NewGomegaWithT(t)

	irqSmpAffinityProcFile := "/tmp/default_smp_affinity"
	irqBalanceConfigFile := "/tmp/irqbalance"

	fa, err := os.OpenFile(irqSmpAffinityProcFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	fa.Write([]byte("00ffffff,ffffffff"))
	fa.Close()
	defer os.Remove(irqSmpAffinityProcFile)

	fi, err := os.OpenFile("/tmp/irqbalance", os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	fi.Close()
	defer os.Remove(irqBalanceConfigFile)

	err = SetIRQLoadBalancing("1-2", false, irqSmpAffinityProcFile, irqBalanceConfigFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", false, irqSmpAffinityProcFile, irqBalanceConfigFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("1-2", true, irqSmpAffinityProcFile, irqBalanceConfigFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", true, irqSmpAffinityProcFile, irqBalanceConfigFile)
	g.Expect(err).NotTo(HaveOccurred())

	fa, err = os.OpenFile(irqSmpAffinityProcFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err := ioutil.ReadAll(fa)
	fa.Close()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("00ffffff,ffffffff"))

	fb, err := os.OpenFile(irqBalanceConfigFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err = ioutil.ReadAll(fb)
	fb.Close()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("ff000000,00000000"))
}
