// Package config 管理应用程序配置
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用程序配置
type Config struct {
	RecentProjects []string `json:"recent_projects"` // 最近打开的项目列表
}

var (
	configPath string
	appConfig  *Config
)

// 初始化配置路径
func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	configDir := filepath.Join(homeDir, ".chisknife")
	os.MkdirAll(configDir, 0755)
	configPath = filepath.Join(configDir, "config.json")
}

// Load 加载配置
func Load() *Config {
	if appConfig != nil {
		return appConfig
	}

	appConfig = &Config{
		RecentProjects: []string{},
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return appConfig
	}

	json.Unmarshal(data, appConfig)
	return appConfig
}

// Save 保存配置
func Save() error {
	if appConfig == nil {
		return nil
	}

	data, err := json.MarshalIndent(appConfig, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// AddRecentProject 添加最近打开的项目
func AddRecentProject(path string) {
	if appConfig == nil {
		appConfig = Load()
	}

	// 移除已存在的相同路径
	for i := 0; i < len(appConfig.RecentProjects); i++ {
		if appConfig.RecentProjects[i] == path {
			appConfig.RecentProjects = append(
				appConfig.RecentProjects[:i],
				appConfig.RecentProjects[i+1:]...,
			)
			break
		}
	}

	// 添加到列表开头
	appConfig.RecentProjects = append([]string{path}, appConfig.RecentProjects...)

	// 限制最多10个
	if len(appConfig.RecentProjects) > 10 {
		appConfig.RecentProjects = appConfig.RecentProjects[:10]
	}

	Save()
}

// GetRecentProjects 获取最近打开的项目列表（过滤不存在的文件）
func GetRecentProjects() []string {
	if appConfig == nil {
		appConfig = Load()
	}

	validProjects := []string{}
	modified := false

	for _, path := range appConfig.RecentProjects {
		if _, err := os.Stat(path); err == nil {
			validProjects = append(validProjects, path)
		} else {
			modified = true
		}
	}

	if modified {
		appConfig.RecentProjects = validProjects
		Save()
	}

	return validProjects
}

// GetLastProject 获取最后打开的项目
func GetLastProject() string {
	projects := GetRecentProjects()
	if len(projects) > 0 {
		return projects[0]
	}
	return ""
}
