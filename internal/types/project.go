// Package types 定义了应用程序使用的核心数据类型
// 包括构建选项、ROM 列表等配置结构
package types

// 定义了卡带构建的配置选项
type Options struct {
	CartridgeType     int32 `json:"cartridge_type"`      // 卡带类型索引
	MinimalRomSize    int32 `json:"minimal_rom_size"`    // 最小 ROM 大小索引
	HaveBattery       bool  `json:"have_battery"`        // 是否包含电池
	UseRTS            bool  `json:"use_rts"`             // 是否使用 RTS
	SplitROM          bool  `json:"split_rom"`           // 是否分割 ROM
	Sram1MSaveSupport bool  `json:"sram1m_save_support"` // 是否支持 1M SRAM 存档
}

// ROM 文件
type Rom struct {
	Name string `json:"name"` // 文件名
	Path string `json:"path"` // 文件路径
}

// ROM 文件路径列表
type RomList []Rom

// 打包工程
type Project struct {
	Options Options `json:"options"` // 卡带选项
	Roms    RomList `json:"roms"`    // rom列表
}
