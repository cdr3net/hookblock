package blocks

import "reflect"

type HBlockFactory func() Block

func HBlockRegistry() map[string]HBlockFactory {
	return map[string]HBlockFactory{
		"http_server":      func() Block { return &HttpServer{} },
		"dead_mans_switch": func() Block { return &DeadMansSwitch{} },
	}
}

type HConfig struct {
	HttpServers      []*HttpServer     `hcl:"http_server,block"`
	DeadMansSwitches []*DeadMansSwitch `hcl:"dead_mans_switch,block"`
}

func (conf *HConfig) AllBlocks() []Block {
	var ret []Block
	v := reflect.ValueOf(conf).Elem()
	numFields := v.NumField()
	for i := 0; i < numFields; i++ {
		fv := v.Field(i)
		aLen := fv.Len()
		for j := 0; j < aLen; j++ {
			ret = append(ret, fv.Index(j).Interface().(Block))
		}
	}
	return ret
}
