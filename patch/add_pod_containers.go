package patch

import (
	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/operation"
	core_v1 "k8s.io/api/core/v1"
)

func AddPodContainers(pod *core_v1.Pod, containers []core_v1.Container) (json_patch.Patch, error) {
	res := make(json_patch.Patch, 0, len(containers))

	notEmpty := len(pod.Spec.Containers) > 0
	for _, c := range containers {
		var (
			op  json_patch.Operation
			err error
		)

		if notEmpty {
			op, err = operation.Add("/spec/containers/-", c)
		} else {
			notEmpty = true
			op, err = operation.Add("/spec/containers", []core_v1.Container{c})
		}

		if err != nil {
			return nil, err
		}
		res = append(res, op)
	}

	return res, nil
}
