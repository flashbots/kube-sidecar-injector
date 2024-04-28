package patch

import (
	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/operation"
	core_v1 "k8s.io/api/core/v1"
)

func UpdatePodLabels(
	pod *core_v1.Pod,
	labels map[string]string,
) (json_patch.Patch, error) {
	if len(labels) == 0 {
		return nil, nil
	}

	if len(pod.Labels) == 0 {
		op, err := operation.Add("/metadata/labels", labels)
		if err != nil {
			return nil, err
		}
		return []json_patch.Operation{op}, nil
	}

	res := make(json_patch.Patch, 0, len(labels))

	for k, v := range labels {
		if o, exists := pod.Labels[k]; exists {
			if o != v {
				op, err := operation.Replace("/metadata/labels/"+operation.Escape(k), v)
				if err != nil {
					return nil, err
				}
				res = append(res, op)
			}
		} else {
			op, err := operation.Add("/metadata/labels/"+operation.Escape(k), v)
			if err != nil {
				return nil, err
			}
			res = append(res, op)
		}
	}

	return res, nil
}
