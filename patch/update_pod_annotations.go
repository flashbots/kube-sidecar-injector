package patch

import (
	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/operation"
	core_v1 "k8s.io/api/core/v1"
)

func UpdatePodAnnotations(
	pod *core_v1.Pod,
	annotations map[string]string,
) (json_patch.Patch, error) {
	if len(pod.Annotations) == 0 {
		op, err := operation.Add("/metadata/annotations", annotations)
		if err != nil {
			return nil, err
		}
		return []json_patch.Operation{op}, nil
	}

	res := make(json_patch.Patch, 0, len(annotations))

	for k, v := range annotations {
		if _, exists := pod.Annotations[k]; exists {
			op, err := operation.Replace("/metadata/annotations/"+operation.Escape(k), v)
			if err != nil {
				return nil, err
			}
			res = append(res, op)
		} else {
			op, err := operation.Add("/metadata/annotations/"+operation.Escape(k), v)
			if err != nil {
				return nil, err
			}
			res = append(res, op)
		}
	}

	return res, nil
}
