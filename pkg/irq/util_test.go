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

	fa, err := os.OpenFile(irqSmpAffinityProcFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	fa.Write([]byte("00ffffff,ffffffff"))
	fa.Close()
	defer os.Remove(irqSmpAffinityProcFile)

	err = SetIRQLoadBalancing("1-2", false, irqSmpAffinityProcFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", false, irqSmpAffinityProcFile)
	g.Expect(err).NotTo(HaveOccurred())

	fa, err = os.OpenFile(irqSmpAffinityProcFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err := ioutil.ReadAll(fa)
	fa.Close()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("00ffffff,ffffffe0"))
}

func TestResetIRQLoadBalancing(t *testing.T) {
	g := NewGomegaWithT(t)

	irqSmpAffinityProcFile := "/tmp/default_smp_affinity"

	fa, err := os.OpenFile(irqSmpAffinityProcFile, os.O_CREATE|os.O_WRONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	fa.Write([]byte("00ffffff,ffffffff"))
	fa.Close()
	defer os.Remove(irqSmpAffinityProcFile)

	err = SetIRQLoadBalancing("1-2", false, irqSmpAffinityProcFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", false, irqSmpAffinityProcFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("1-2", true, irqSmpAffinityProcFile)
	g.Expect(err).NotTo(HaveOccurred())

	err = SetIRQLoadBalancing("0,3-4", true, irqSmpAffinityProcFile)
	g.Expect(err).NotTo(HaveOccurred())

	fa, err = os.OpenFile(irqSmpAffinityProcFile, os.O_RDONLY, 0644)
	g.Expect(err).NotTo(HaveOccurred())
	rawBytes, err := ioutil.ReadAll(fa)
	fa.Close()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(rawBytes)).To(Equal("00ffffff,ffffffff"))
}
