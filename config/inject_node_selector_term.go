package config

import (
	"hash"

	core_v1 "k8s.io/api/core/v1"
)

type InjectNodeSelectorTerm struct {
	MatchExpressions []InjectMatchExpression `yaml:"matchExpressions,omitempty"`
	MatchFields      []InjectMatchExpression `yaml:"matchFields,omitempty"`
}

func (nst InjectNodeSelectorTerm) hash(sum hash.Hash64) {
	{ // matchExpressions
		if len(nst.MatchExpressions) > 0 {
			sum.Write([]byte("matchExpressions:"))
			for _, me := range nst.MatchExpressions {
				me.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}

	{ // matchFields
		if len(nst.MatchFields) > 0 {
			sum.Write([]byte("matchFields:"))
			for _, me := range nst.MatchFields {
				me.hash(sum)
			}
			sum.Write([]byte{255})
		}
	}
}

func (nst InjectNodeSelectorTerm) NodeSelectorTerm() (*core_v1.NodeSelectorTerm, error) {
	nodeSelectorTerm := &core_v1.NodeSelectorTerm{}

	if len(nst.MatchExpressions) > 0 {
		matchExpressions := make([]core_v1.NodeSelectorRequirement, 0, len(nst.MatchExpressions))
		for _, me := range nst.MatchExpressions {
			matchExpressions = append(matchExpressions, core_v1.NodeSelectorRequirement{
				Key:      me.Key,
				Operator: core_v1.NodeSelectorOperator(me.Operator),
				Values:   me.Values,
			})
		}
		nodeSelectorTerm.MatchExpressions = matchExpressions
	}

	if len(nst.MatchFields) > 0 {
		matchFields := make([]core_v1.NodeSelectorRequirement, 0, len(nst.MatchFields))
		for _, mf := range nst.MatchExpressions {
			matchFields = append(matchFields, core_v1.NodeSelectorRequirement{
				Key:      mf.Key,
				Operator: core_v1.NodeSelectorOperator(mf.Operator),
				Values:   mf.Values,
			})
		}
		nodeSelectorTerm.MatchFields = matchFields
	}

	return nodeSelectorTerm, nil
}
