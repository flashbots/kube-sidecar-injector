package config

import (
	"hash"

	core_v1 "k8s.io/api/core/v1"
)

type InjectVolume struct {
	Name string `yaml:"name,omitempty"`

	ConfigMap *InjectVolumeConfigMap `yaml:"configMap,omitempty"`
}

func (v InjectVolume) hash(sum hash.Hash64) {
	sum.Write([]byte("name:"))
	sum.Write([]byte(v.Name))
	sum.Write([]byte{255})

	if v.ConfigMap != nil {
		sum.Write([]byte("configMap:"))
		v.ConfigMap.hash(sum)
		sum.Write([]byte{255})
	}
}

func (v InjectVolume) Volume() (*core_v1.Volume, error) {
	res := &core_v1.Volume{
		Name:         v.Name,
		VolumeSource: core_v1.VolumeSource{},
	}

	{ // configMap
		if v.ConfigMap != nil {
			configMap, err := v.ConfigMap.ConfigMapVolumeSource()
			if err != nil {
				return nil, err
			}
			res.VolumeSource.ConfigMap = configMap
		}
	}

	return res, nil
}
