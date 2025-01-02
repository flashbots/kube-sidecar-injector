package patch

import (
	json_patch "github.com/evanphx/json-patch"
	"github.com/flashbots/kube-sidecar-injector/config"
	"github.com/flashbots/kube-sidecar-injector/operation"
	core_v1 "k8s.io/api/core/v1"
)

func InsertAffinity(
	pod *core_v1.Pod,
	affinity *config.InjectAffinity,
) (json_patch.Patch, error) {
	if affinity == nil {
		return nil, nil
	}

	if pod.Spec.Affinity != nil &&
		pod.Spec.Affinity.NodeAffinity != nil &&
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil &&
		len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) > 0 {
		// do not patch pods with already defined affinity
		return nil, nil
	}

	_affinity, err := affinity.Affinity()
	if err != nil {
		return nil, err
	}

	op, err := operation.Add("/spec/affinity", _affinity)
	if err != nil {
		return nil, err
	}

	return json_patch.Patch{op}, nil
}
