package config

import (
	"hash"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InjectLabelSelector struct {
	MatchLabels      map[string]string       `yaml:"matchLabels,omitempty"`
	MatchExpressions []InjectMatchExpression `yaml:"matchExpressions,omitempty"`
}

func (ls InjectLabelSelector) hash(sum hash.Hash64) {
	{ // matchLabels
		if len(ls.MatchLabels) > 0 {
			sum.Write([]byte("matchLabels:"))
			for k, v := range ls.MatchLabels {
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

	{ // matchExpressions
		if len(ls.MatchExpressions) > 0 {
			sum.Write([]byte("matchExpressions:"))
			for _, me := range ls.MatchExpressions {
				me.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}
}

func (ls InjectLabelSelector) LabelSelector() (*meta_v1.LabelSelector, error) {
	matchExpressions := make([]meta_v1.LabelSelectorRequirement, 0, len(ls.MatchExpressions))
	for _, me := range ls.MatchExpressions {
		matchExpressions = append(matchExpressions, meta_v1.LabelSelectorRequirement{
			Key:      me.Key,
			Operator: meta_v1.LabelSelectorOperator(me.Operator),
			Values:   me.Values,
		})
	}

	return &meta_v1.LabelSelector{
		MatchLabels:      ls.MatchLabels,
		MatchExpressions: matchExpressions,
	}, nil
}
