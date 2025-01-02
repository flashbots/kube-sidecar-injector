package config

import (
	"hash"

	core_v1 "k8s.io/api/core/v1"
)

type InjectNodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution *InjectRequiredDuringSchedulingIgnoredDuringExecution `yaml:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

func (na InjectNodeAffinity) hash(sum hash.Hash64) {
	{ // requiredDuringSchedulingIgnoredDuringExecution
		if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			sum.Write([]byte("requiredDuringSchedulingIgnoredDuringExecution:"))
			na.RequiredDuringSchedulingIgnoredDuringExecution.hash(sum)
			sum.Write([]byte{255})
		}
	}
}

func (na InjectNodeAffinity) NodeAffinity() (*core_v1.NodeAffinity, error) {
	nodeAffinity := &core_v1.NodeAffinity{}

	if na.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		if len(na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) > 0 {
			nodeSelectorTerms := make(
				[]core_v1.NodeSelectorTerm,
				0,
				len(na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms),
			)

			for _, nst := range na.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				_nst, err := nst.NodeSelectorTerm()
				if err != nil {
					return nil, err
				}
				nodeSelectorTerms = append(nodeSelectorTerms, *_nst)
			}

			nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &core_v1.NodeSelector{
				NodeSelectorTerms: nodeSelectorTerms,
			}
		}
	}

	return nodeAffinity, nil
}
