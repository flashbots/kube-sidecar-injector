package config

import (
	"hash"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabelSelector struct {
	MatchLabels      map[string]string `yaml:"matchLabels,omitempty"`
	MatchExpressions []MatchExpression `yaml:"matchExpressions,omitempty"`
}

type MatchExpression struct {
	Key      string   `yaml:"key,omitempty"`
	Operator string   `yaml:"operator,omitempty"`
	Values   []string `yaml:"values,omitempty"`
}

func (l LabelSelector) hash(sum hash.Hash64) {
	sum.Write([]byte("matchLabels:"))
	for k, v := range l.MatchLabels {
		sum.Write([]byte("key:"))
		sum.Write([]byte(k))
		sum.Write([]byte{255})

		sum.Write([]byte("value:"))
		sum.Write([]byte(v))
		sum.Write([]byte{255})
	}

	sum.Write([]byte("matchExpressions:"))
	for _, m := range l.MatchExpressions {
		sum.Write([]byte("key:"))
		sum.Write([]byte(m.Key))
		sum.Write([]byte{255})

		sum.Write([]byte("operator:"))
		sum.Write([]byte(m.Operator))
		sum.Write([]byte{255})

		sum.Write([]byte("values:"))
		for _, v := range m.Values {
			sum.Write([]byte(v))
			sum.Write([]byte{255})
		}
	}
}

func (l LabelSelector) LabelSelector() (*meta_v1.LabelSelector, error) {
	matchExpressions := make([]meta_v1.LabelSelectorRequirement, 0, len(l.MatchExpressions))
	for _, m := range l.MatchExpressions {
		matchExpressions = append(matchExpressions, meta_v1.LabelSelectorRequirement{
			Key:      m.Key,
			Operator: meta_v1.LabelSelectorOperator(m.Operator),
			Values:   m.Values,
		})
	}

	return &meta_v1.LabelSelector{
		MatchLabels:      l.MatchLabels,
		MatchExpressions: matchExpressions,
	}, nil
}
