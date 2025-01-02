package config

import (
	"fmt"
	"hash/fnv"
)

type Inject struct {
	Name string `yaml:"name,omitempty"`

	MaxIterations int `yaml:"maxIterations,omitempty"`

	LabelSelector     *InjectLabelSelector `yaml:"labelSelector,omitempty"`
	NamespaceSelector *InjectLabelSelector `yaml:"namespaceSelector,omitempty"`

	Labels map[string]string `yaml:"labels,omitempty"`

	Affinity     *InjectAffinity     `yaml:"affinity,omitempty"`
	Containers   []InjectContainer   `yaml:"containers,omitempty"`
	Tolerations  []InjectToleration  `yaml:"tolerations,omitempty"`
	VolumeMounts []InjectVolumeMount `yaml:"volumeMounts,omitempty"`
	Volumes      []InjectVolume      `yaml:"volumes,omitempty"`
}

func (i Inject) Fingerprint() string {
	sum := fnv.New64a()

	{ // name
		sum.Write([]byte("name:"))
		sum.Write([]byte(i.Name))
		sum.Write([]byte{255})
	}

	{ // labelSelector
		if i.LabelSelector != nil {
			sum.Write([]byte("labelSelector:"))
			i.LabelSelector.hash(sum)
			sum.Write([]byte{255})
		}
	}

	{ // namespaceSelector
		if i.NamespaceSelector != nil {
			sum.Write([]byte("namespaceSelector:"))
			i.NamespaceSelector.hash(sum)
			sum.Write([]byte{255})
		}
	}

	{ // labels
		if len(i.Labels) > 0 {
			sum.Write([]byte("labels:"))
			for k, v := range i.Labels {
				sum.Write([]byte("key:"))
				sum.Write([]byte(k))
				sum.Write([]byte{255})

				sum.Write([]byte("value:"))
				sum.Write([]byte(v))
				sum.Write([]byte{255})
			}
			sum.Write([]byte{255})
		}
	}

	{ // affinity
		if i.Affinity != nil {
			sum.Write([]byte("affinity:"))
			i.Affinity.hash(sum)
			sum.Write([]byte{255})
		}
	}

	{ // containers
		if len(i.Containers) > 0 {
			sum.Write([]byte("containers:"))
			for _, c := range i.Containers {
				c.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}

	{ // tolerations
		if len(i.Tolerations) > 0 {
			sum.Write([]byte("tolerations:"))
			for _, t := range i.Tolerations {
				t.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}

	{ // volumeMounts
		if len(i.VolumeMounts) > 0 {
			sum.Write([]byte("volumeMounts:"))
			for _, vm := range i.VolumeMounts {
				vm.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}

	{ // volumes
		if len(i.Volumes) > 0 {
			sum.Write([]byte("volumes:"))
			for _, v := range i.Volumes {
				v.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}

	return fmt.Sprintf("%016x", sum.Sum64())
}
