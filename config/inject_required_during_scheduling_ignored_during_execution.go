package config

import "hash"

type InjectRequiredDuringSchedulingIgnoredDuringExecution struct {
	NodeSelectorTerms []InjectNodeSelectorTerm `yaml:"nodeSelectorTerms"`
}

func (rdside InjectRequiredDuringSchedulingIgnoredDuringExecution) hash(sum hash.Hash64) {
	{ // nodeSelectorTerms
		if len(rdside.NodeSelectorTerms) > 0 {
			sum.Write([]byte("requiredDuringSchedulingIgnoredDuringExecution:"))
			for _, nst := range rdside.NodeSelectorTerms {
				nst.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}
}
