package ctyutil

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func Convert(v interface{}) (cty.Value, error) {
	var err error

	if vv, ok := v.(map[string]interface{}); ok {
		ret := make(map[string]cty.Value)
		for k, v := range vv {
			ret[k], err = Convert(v)
			if err != nil {
				return cty.Value{}, err
			}
		}
		return cty.ObjectVal(ret), nil
	} else if vv, ok := v.([]interface{}); ok {
		var ret []cty.Value
		for _, v := range vv {
			r, err := Convert(v)
			if err != nil {
				return cty.Value{}, err
			}
			ret = append(ret, r)
		}
		return cty.TupleVal(ret), nil
	} else {
		it, err := gocty.ImpliedType(v)
		if err != nil {
			return cty.Value{}, err
		}
		value, err := gocty.ToCtyValue(v, it)
		if err != nil {
			return cty.Value{}, err
		}
		return value, nil
	}
}

func StrMapValue(m map[string]string) cty.Value {
	ret := make(map[string]cty.Value)
	for k, v := range m {
		ret[k] = cty.StringVal(v)
	}
	return cty.MapVal(ret)
}
