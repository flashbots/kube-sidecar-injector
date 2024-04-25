package config

import (
	"fmt"
	"hash/fnv"
)

type Inject struct {
	LabelSelector *LabelSelector `yaml:"labelSelector,omitempty"`

	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`

	Containers []Container `yaml:"containers,omitempty"`
}

func (i Inject) Fingerprint() string {
	sum := fnv.New64a()

	sum.Write([]byte("labelSelector:"))
	i.LabelSelector.hash(sum)

	sum.Write([]byte("annotations:"))
	for k, v := range i.Annotations {
		sum.Write([]byte("key:"))
		sum.Write([]byte(k))
		sum.Write([]byte{255})

		sum.Write([]byte("value:"))
		sum.Write([]byte(v))
		sum.Write([]byte{255})
	}

	sum.Write([]byte("labels:"))
	for k, v := range i.Labels {
		sum.Write([]byte("key:"))
		sum.Write([]byte(k))
		sum.Write([]byte{255})

		sum.Write([]byte("value:"))
		sum.Write([]byte(v))
		sum.Write([]byte{255})
	}

	sum.Write([]byte("containers:"))
	for _, c := range i.Containers {
		c.hash(sum)
	}

	return fmt.Sprintf("%016x", sum.Sum64())
}
