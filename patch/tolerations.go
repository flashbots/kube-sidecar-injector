package patch

import (
	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/operation"
	core_v1 "k8s.io/api/core/v1"
)

func InsertTolerations(
	pod *core_v1.Pod,
	tolerations []core_v1.Toleration,
) (json_patch.Patch, error) {
	if len(tolerations) == 0 {
		return nil, nil
	}

	res := make(json_patch.Patch, 0, len(tolerations))

	notEmpty := len(pod.Spec.Tolerations) > 0
	for _, t := range tolerations {
		var (
			op  json_patch.Operation
			err error
		)

		if notEmpty {
			op, err = operation.Add("/spec/tolerations/-", t)
		} else {
			notEmpty = true
			op, err = operation.Add("/spec/tolerations", []core_v1.Toleration{t})
		}

		if err != nil {
			return nil, err
		}
		res = append(res, op)
	}

	return res, nil
}
