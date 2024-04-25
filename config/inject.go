package config

import (
	"fmt"
	"hash/fnv"
)

type Inject struct {
	LabelSelector *LabelSelector `yaml:"labelSelector,omitempty"`
	Containers    []Container    `yaml:"containers,omitempty"`
}

func (i Inject) Fingerprint() string {
	sum := fnv.New64a()

	sum.Write([]byte("labelSelector:"))
	i.LabelSelector.hash(sum)

	sum.Write([]byte("containers:"))
	for _, c := range i.Containers {
		c.hash(sum)
	}

	return fmt.Sprintf("%016x", sum.Sum64())
}
