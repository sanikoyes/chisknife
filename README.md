# ChisKnife2

基于 giu (imgui) 的 ChisFlash 瑞士军刀工具。

## 项目结构

```
chisknife/
├── asset/              # 资源文件
│   ├── asset.go       # 嵌入资源
│   └── bg.png         # 背景图片
├── internal/
│   ├── preset/        # 预设数据
│   │   ├── cartridge_types.go
│   │   └── rom_sizes.go
│   ├── types/         # 类型定义
│   │   ├── build_option.go
│   │   ├── cartridge_type.go
│   │   └── rom_size.go
│   └── ui/            # UI 界面
│       ├── cart_settings.go
│       ├── main_window.go
│       └── rom_list.go
├── translations/      # 国际化翻译
│   ├── en.json
│   └── zh-CN.json
├── main.go           # 主程序入口
├── go.mod            # Go 模块定义
└── build.sh          # 构建脚本

```

## 依赖

- Go 1.25.0+
- giu (github.com/AllenDang/giu)

## 构建

```bash
# 安装依赖
go mod tidy

# 构建
./build.sh

# 或直接运行
go run .
```

## 功能

- ROM 列表管理
- 卡带设置配置
- 菜单背景自定义
- 多语言支持（中文/英文）

## 与 chisknife 的区别

- UI 框架从 Fyne 改为 giu (基于 imgui)
- 保持相同的项目结构和功能逻辑
- 使用 imgui 的即时模式 GUI 渲染
