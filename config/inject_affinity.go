package config

import (
	"hash"

	core_v1 "k8s.io/api/core/v1"
)

type InjectAffinity struct {
	NodeAffinity *InjectNodeAffinity `yaml:"nodeAffinity,omitempty"`
}

func (a InjectAffinity) hash(sum hash.Hash64) {
	{ // nodeAffinity
		if a.NodeAffinity != nil {
			sum.Write([]byte("nodeAffinity:"))
			a.NodeAffinity.hash(sum)
			sum.Write([]byte{255})
		}
	}
}

func (a InjectAffinity) Affinity() (*core_v1.Affinity, error) {
	var (
		nodeAffinity *core_v1.NodeAffinity
		err          error
	)

	if a.NodeAffinity != nil {
		nodeAffinity, err = a.NodeAffinity.NodeAffinity()
		if err != nil {
			return nil, err
		}
	}

	return &core_v1.Affinity{
		NodeAffinity: nodeAffinity,
	}, nil
}
