package types

type CartridgeType struct {
	Name       string
	FlashSize  int
	SectorSize int
	BlockSize  int
}

type CartridgeTypes []CartridgeType

func (t CartridgeTypes) Names() []string {
	names := make([]string, len(t))
	for i, v := range t {
		names[i] = v.Name
	}
	return names
}
