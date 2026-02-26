// Package preset 提供预定义的卡带类型和 ROM 大小配置
// 包含常见的闪存芯片规格和参数
package preset

import "chisknife/internal/types"

// 定义了支持的卡带类型列表
// 每种类型包含不同的闪存芯片规格，如容量、扇区大小和块大小
var CartridgeTypes = types.CartridgeTypes{
	{
		Name:       "MSP55LV100S or S29GL512",
		FlashSize:  0x4000000,
		SectorSize: 0x20000,
		BlockSize:  0x80000,
	},
	{
		Name:       "6600M0U0BE",
		FlashSize:  0x10000000,
		SectorSize: 0x40000,
		BlockSize:  0x80000,
	},
	{
		Name:       "MSP54LV100 or S29GL01G",
		FlashSize:  0x8000000,
		SectorSize: 0x20000,
		BlockSize:  0x80000,
	},
	{
		Name:       "F0095H0",
		FlashSize:  0x20000000,
		SectorSize: 0x40000,
		BlockSize:  0x80000,
	},
	{
		Name:       "S70GL02G",
		FlashSize:  0x10000000,
		SectorSize: 0x20000,
		BlockSize:  0x80000,
	},
}
