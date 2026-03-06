// Package menu 提供 GBA 多游戏菜单构建功能
package menu

// 构建信息
type BuildInfo struct {
	Path    string // 文件路径
	Type    string // 操作类型
	Message string // 消息
	Success bool   // 是否成功
}

// 游戏配置
type Game struct {
	Enabled   bool     `json:"enabled"`
	File      string   `json:"file"`
	Title     string   `json:"title"`
	TitleFont int      `json:"title_font"`
	SaveSlot  *int     `json:"save_slot"` // 使用指针以支持 null
	Keys      []string `json:"keys,omitempty"`
	Map256M   bool     `json:"map_256m,omitempty"`
}

// 卡带配置
type CartridgeConfig struct {
	Type           int  `json:"type"`
	BatteryPresent bool `json:"battery_present"`
	MinRomSize     int  `json:"min_rom_size"`
}

// 完整配置
type Config struct {
	Cartridge CartridgeConfig `json:"cartridge"`
	Games     []Game          `json:"games"`
}

// 构建选项
type BuildOptions struct {
	// 卡带选项
	CartridgeType  int  // 卡带类型 (1-5)
	BatteryPresent bool // 是否有电池
	MinRomSize     int  // 最小 ROM 大小

	// 补丁选项
	SRAMBankType        byte // SRAM bank 类型
	BatterylessAutoSave bool // Batteryless 自动保存
	UseRTS              bool // 使用 RTS

	// 输出选项
	ConfigPath  string // 配置文件路径
	RomBasePath string // ROM 基础路径
	OutputPath  string // 输出文件路径
	BgPath      string // 背景图片路径
	Split       bool   // 是否分割输出
}

// 游戏输入信息
type GameInput struct {
	Path     string // ROM 文件路径
	Name     string // 游戏名称
	SaveSlot *int   // 存档槽位（nil 表示无存档）
}

// 构建参数（用于 MenuBuilder）
type BuildArgs struct {
	Options  BuildOptions // 构建选项
	GameList []GameInput  // 游戏列表
}
