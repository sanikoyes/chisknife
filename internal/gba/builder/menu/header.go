package menu

import (
	"os"
	"strings"
)

// 从 GBA ROM 读取游戏 ID
func GetROMID(romPath string) (string, error) {
	f, err := os.Open(romPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 跳转到 0xAC 位置
	_, err = f.Seek(0xAC, 0)
	if err != nil {
		return "", err
	}

	// 读取 4 字节 ID
	id := make([]byte, 4)
	_, err = f.Read(id)
	if err != nil {
		return "", err
	}

	return string(id), nil
}

// 从 GBA ROM 读取游戏名称
func GetROMName(romPath string) (string, error) {
	f, err := os.Open(romPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// 跳转到 0xA0 位置
	_, err = f.Seek(0xA0, 0)
	if err != nil {
		return "", err
	}

	// 读取 12 字节名称
	name := make([]byte, 12)
	_, err = f.Read(name)
	if err != nil {
		return "", err
	}

	// 移除空字节并去除前导空格
	nameStr := strings.ReplaceAll(string(name), "\x00", " ")
	return strings.TrimLeft(nameStr, " "), nil
}

// 从 GBA ROM 读取版本号
func GetROMVersion(romPath string) (byte, error) {
	f, err := os.Open(romPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// 跳转到 0xBC 位置
	_, err = f.Seek(0xBC, 0)
	if err != nil {
		return 0, err
	}

	// 读取 1 字节版本
	version := make([]byte, 1)
	_, err = f.Read(version)
	if err != nil {
		return 0, err
	}

	return version[0], nil
}
