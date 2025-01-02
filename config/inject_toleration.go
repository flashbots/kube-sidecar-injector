package config

import (
	"hash"
	"unsafe"

	core_v1 "k8s.io/api/core/v1"
)

type InjectToleration struct {
	Key               string `yaml:"key,omitempty"`
	Operator          string `yaml:"operator,omitempty"`
	Value             string `yaml:"value,omitempty"`
	Effect            string `yaml:"effect,omitempty"`
	TolerationSeconds *int64 `yaml:"tolerationSeconds,omitempty"`
}

func (t InjectToleration) hash(sum hash.Hash64) {
	{ // key
		sum.Write([]byte("key:"))
		sum.Write([]byte(t.Key))
		sum.Write([]byte{255})
	}

	{ // operator
		sum.Write([]byte("operator:"))
		sum.Write([]byte(t.Operator))
		sum.Write([]byte{255})
	}

	{ // value
		sum.Write([]byte("value:"))
		sum.Write([]byte(t.Value))
		sum.Write([]byte{255})
	}

	{ // effect
		sum.Write([]byte("effect:"))
		sum.Write([]byte(t.Effect))
		sum.Write([]byte{255})
	}

	{ // tolerationSeconds
		if t.TolerationSeconds != nil {
			sum.Write([]byte("tolerationSeconds:"))
			sum.Write(unsafe.Slice(
				(*byte)(unsafe.Pointer(t.TolerationSeconds)),
				unsafe.Sizeof(*t.TolerationSeconds),
			))
			sum.Write([]byte{255})
		}
	}
}

func (t InjectToleration) Toleration() (*core_v1.Toleration, error) {
	return &core_v1.Toleration{
		Key:               t.Key,
		Operator:          core_v1.TolerationOperator(t.Operator),
		Value:             t.Value,
		Effect:            core_v1.TaintEffect(t.Effect),
		TolerationSeconds: t.TolerationSeconds,
	}, nil
}
