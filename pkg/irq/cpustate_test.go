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
	"testing"

	. "github.com/onsi/gomega"
)

const (
	c1AssignedCPUs = "1-2,29"
	c2AssignedCPUs = "3-5,30"
	c3AssignedCPUs = "6-7"
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
	cacheV1["8631b3ef-066d-4723-a4b2-797d9d095c4f"] = c1AssignedCPUs
	cms, err := NewCPUManagerServiceWithEntries(cacheV1, nil)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal(c1AssignedCPUs))
	cms.Remove("8631b3ef-066d-4723-a4b2-797d9d095c4f")
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal(""))
}

func TestNewCPUManagerV2(t *testing.T) {
	g := NewGomegaWithT(t)
	cacheV2 := make(map[string]map[string]string)
	cacheV2["8631b3ef-066d-4723-a4b2-797d9d095c4f"] = make(map[string]string)
	cacheV2["8631b3ef-066d-4723-a4b2-797d9d095c4f"]["busybox"] = c1AssignedCPUs
	cacheV2["9631b3ef-066d-4723-a4b2-797d9d095c50"] = make(map[string]string)
	cacheV2["9631b3ef-066d-4723-a4b2-797d9d095c50"]["busybox1"] = c2AssignedCPUs
	cacheV2["9631b3ef-066d-4723-a4b2-797d9d095c50"]["busybox2"] = c3AssignedCPUs
	cms, err := NewCPUManagerServiceWithEntries(nil, cacheV2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal(c1AssignedCPUs))
	cms.Remove("8631b3ef-066d-4723-a4b2-797d9d095c4f")
	g.Expect(cms.GetAssignedCpusFromCache("8631b3ef-066d-4723-a4b2-797d9d095c4f")).To(Equal(""))
	g.Expect(cms.GetAssignedCpusFromCache("9631b3ef-066d-4723-a4b2-797d9d095c50")).To(ContainSubstring(c2AssignedCPUs))
	g.Expect(cms.GetAssignedCpusFromCache("9631b3ef-066d-4723-a4b2-797d9d095c50")).To(ContainSubstring(c3AssignedCPUs))
}
