package blocks

type DeadMansSwitch struct {
	Name string `hcl:"name,label"`
}

func (b DeadMansSwitch) GetName() string {
	return b.Name
}
