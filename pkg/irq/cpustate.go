package irq

import (
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/state"
)

const (
	kubeletRootDir          string = "/shared/var/lib/kubelet/"
	cpuManagerStateFileName string = "cpu_manager_state"
)

// GetAssignedCpusforPod get allocated cpu cores for given Guaranteed QoS pod uid
func GetAssignedCpusforPod(podUID string) ([]int, error) {
	_, err := checkpointmanager.NewCheckpointManager(kubeletRootDir)
	if err != nil {
		return nil, err
	}
	//TODO
	return nil, nil
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

func restoreState(cm checkpointmanager.CheckpointManager) error {

	var err error

	checkpointV1 := newCPUManagerCheckpointV1()
	checkpointV2 := newCPUManagerCheckpointV2()

	if err = cm.GetCheckpoint(cpuManagerStateFileName, checkpointV1); err != nil {
		checkpointV1 = &state.CPUManagerCheckpointV1{} // reset it back to 0
		if err = cm.GetCheckpoint(cpuManagerStateFileName, checkpointV2); err != nil {
			//TODO
		}
	} else {
		//TODO
	}

	return nil
}

