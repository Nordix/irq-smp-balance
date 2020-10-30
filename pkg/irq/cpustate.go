package irq

import (
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/state"
)

const (
	kubeletRootDir          string = "/shared/var/lib/kubelet/"
	cpuManagerStateFileName string = "cpu_manager_state"
)

// CPUManagerService APIs for retriving assigned cpus
type CPUManagerService interface {
	GetAssignedCpus(podUID string) string
	GetAssignedCpusFromCache(podUID string) string
}

type cpuState struct {
	checkpoint checkpointmanager.CheckpointManager
	EntriesV1  map[string]string
	EntriesV2  map[string]map[string]string
}

// GetAssignedCpus get allocated cpu cores for given Guaranteed QoS pod uid
func (cs *cpuState) GetAssignedCpus(podUID string) string {
	cs.restoreState()
	return cs.GetAssignedCpusFromCache(podUID)
}

// GetAssignedCpus get allocated cpu cores for given Guaranteed QoS pod uid
// can be used in pod delete scenarios
func (cs *cpuState) GetAssignedCpusFromCache(podUID string) string {
	cpus := cs.getCPUsFromCheckpointV1(podUID)
	if cpus != "" {
		return cpus
	}
	return cs.getCPUsFromCheckpointV2(podUID)
}

func (cs *cpuState) getCPUsFromCheckpointV1(podUID string) string {
	return cs.EntriesV1[podUID]
}

func (cs *cpuState) getCPUsFromCheckpointV2(podUID string) string {
	var cpus string
	for _, cpu := range cs.EntriesV2[podUID] {
		cpus = cpus + "," + cpu
	}
	return strings.TrimPrefix(cpus, ",")
}

func newCPUManagerCheckpointV1() *state.CPUManagerCheckpointV1 {
	return &state.CPUManagerCheckpointV1{
		Entries: make(map[string]string),
	}
}

func newCPUManagerCheckpointV2() *state.CPUManagerCheckpointV2 {
	return &state.CPUManagerCheckpointV2{
		Entries: make(map[string]map[string]string),
	}
}

func (cs *cpuState) restoreState() {
	checkpointV1 := newCPUManagerCheckpointV1()
	checkpointV2 := newCPUManagerCheckpointV2()
	if err := cs.checkpoint.GetCheckpoint(cpuManagerStateFileName, checkpointV1); err != nil {
		checkpointV1 = &state.CPUManagerCheckpointV1{}
		if err := cs.checkpoint.GetCheckpoint(cpuManagerStateFileName, checkpointV2); err != nil {
			if checkpointV2.PolicyName != string(cpumanager.PolicyStatic) {
				logrus.Infof("cpu manager policy is not static. no dedicated cpus")
				return
			}
			cs.EntriesV2 = checkpointV2.Entries
		}
	} else {
		if checkpointV1.PolicyName != string(cpumanager.PolicyStatic) {
			logrus.Infof("cpu manager policy is not static. no dedicated cpus")
			return
		}
		cs.EntriesV1 = checkpointV1.Entries
	}
}

// NewCPUManagerService returns new cpu manager service
func NewCPUManagerService() (CPUManagerService, error) {
	cm, err := checkpointmanager.NewCheckpointManager(kubeletRootDir)
	if err != nil {
		return nil, err
	}
	return &cpuState{
		checkpoint: cm,
	}, nil
}

