package types

type Options struct {
	CartridgeType     int32 `json:"cartridge_type"`
	MinimalRomSize    int32 `json:"minimal_rom_size"`
	HaveBattery       bool  `json:"have_battery"`
	UseRTS            bool  `json:"use_rts"`
	SplitROM          bool  `json:"split_rom"`
	Sram1MSaveSupport bool  `json:"sram1m_save_support"`
}

type RomList struct {
	Roms []string `json:"roms"`
}

type BuildOptions struct {
	Options Options `json:"options"`
	RomList RomList `json:"rom_list"`
}
