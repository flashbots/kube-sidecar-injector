package patch

import (
	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/operation"
	core_v1 "k8s.io/api/core/v1"
)

func InsertPodVolumes(
	pod *core_v1.Pod,
	volumes []core_v1.Volume,
) (json_patch.Patch, error) {
	if len(volumes) == 0 {
		return nil, nil
	}

	res := make(json_patch.Patch, 0, len(volumes))

	notEmpty := len(pod.Spec.Volumes) > 0
	for _, v := range volumes {
		var (
			op  json_patch.Operation
			err error
		)

		if notEmpty {
			op, err = operation.Add("/spec/volumes/-", v)
		} else {
			notEmpty = true
			op, err = operation.Add("/spec/volumes", []core_v1.Volume{v})
		}

		if err != nil {
			return nil, err
		}
		res = append(res, op)
	}

	return res, nil
}
