package irq

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestNewCPUManager(t *testing.T) {
	g := NewGomegaWithT(t)
	cms, err := NewCPUManagerService()
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cms.GetAssignedCpusFromCache("test")).To(Equal(""))
}

func TestNewCPUManagerV1(t *testing.T) {
	g := NewGomegaWithT(t)
	_, err := NewCPUManagerServiceWithEntries(nil, nil)
	g.Expect(err).To(HaveOccurred())

	cacheV1 := make(map[string]string)
	cacheV1["8631b3ef-066d-4723-a4b2-797d9d095c4f"] = "1-2,29"
	cms, err := NewCPUManagerServiceWithEntries(cacheV1, nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal("1-2,29"))
	cms.Remove("8631b3ef-066d-4723-a4b2-797d9d095c4f")
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal(""))
}

func TestNewCPUManagerV2(t *testing.T) {
	g := NewGomegaWithT(t)
	cacheV2 := make(map[string]map[string]string)
	cacheV2["8631b3ef-066d-4723-a4b2-797d9d095c4f"] = make(map[string]string)
	cacheV2["8631b3ef-066d-4723-a4b2-797d9d095c4f"]["busybox"] = "1-2,29"
	cacheV2["9631b3ef-066d-4723-a4b2-797d9d095c50"] = make(map[string]string)
	cacheV2["9631b3ef-066d-4723-a4b2-797d9d095c50"]["busybox1"] = "3-5,30"
	cacheV2["9631b3ef-066d-4723-a4b2-797d9d095c50"]["busybox2"] = "6-7"
	cms, err := NewCPUManagerServiceWithEntries(nil, cacheV2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal("1-2,29"))
	cms.Remove("8631b3ef-066d-4723-a4b2-797d9d095c4f")
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal(""))
	g.Expect(cms.GetAssignedCpusFromCache("9631b3ef-066d-4723-a4b2-797d9d095c50")).To(ContainSubstring("3-5,30"))
	g.Expect(cms.GetAssignedCpusFromCache("9631b3ef-066d-4723-a4b2-797d9d095c50")).To(ContainSubstring("6-7"))
}
