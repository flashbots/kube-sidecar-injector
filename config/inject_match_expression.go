package config

import "hash"

type InjectMatchExpression struct {
	Key      string   `yaml:"key,omitempty"`
	Operator string   `yaml:"operator,omitempty"`
	Values   []string `yaml:"values,omitempty"`
}

func (me InjectMatchExpression) hash(sum hash.Hash64) {
	{ // key
		sum.Write([]byte("key:"))
		sum.Write([]byte(me.Key))
		sum.Write([]byte{255})
	}

	{ // operator
		sum.Write([]byte("operator:"))
		sum.Write([]byte(me.Operator))
		sum.Write([]byte{255})
	}

	{ // values
		if len(me.Values) > 0 {
			sum.Write([]byte("values:"))
			for _, v := range me.Values {
				sum.Write([]byte(v))
				sum.Write([]byte{255})
			}
			sum.Write([]byte{255})
		}
	}
}
