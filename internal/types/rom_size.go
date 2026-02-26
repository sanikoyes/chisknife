package types

type RomSize struct {
	Desc string
	Size int
}

type RomSizes []RomSize

func (t RomSizes) Descs() []string {
	descs := make([]string, len(t))
	for i, v := range t {
		descs[i] = v.Desc
	}
	return descs
}
