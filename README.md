# 🚀 SVM - SDK 版本管理器

<div align="center">
  
![SVM Logo](https://img.shields.io/badge/SVM-SDK%20Version%20Manager-blue?style=for-the-badge)
  
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://go.dev/)
[![Made with Cursor](https://img.shields.io/badge/Made%20with-Cursor%20AI-blueviolet)](https://cursor.sh/)

**一个强大的多语言SDK版本管理工具，让你轻松切换各种编程语言的版本**

</div>

## ✨ AI 驱动的开发

> **惊人事实**: 这个项目完全由 [Cursor AI](https://cursor.sh/) 辅助完成，从设计到实现，没有手写一行代码！整个开发过程仅用了 **12 小时**！

作为一个实验性项目，SVM 展示了 AI 辅助编程的强大能力。整个代码库是通过与 Cursor AI 的对话生成的，包括架构设计、功能实现和错误修复。这种开发方式极大地提高了效率，也为未来的软件开发提供了新的可能性。

## 🌟 主要特点

- 🔄 **多语言支持**: 管理 Node.js, Go, Java, Python 等多种语言环境
- 🔍 **版本发现**: 自动获取官方最新版本列表
- 📦 **简单安装**: 一键安装任意版本的SDK
- 🔀 **快速切换**: 在不同版本间无缝切换
- 🔧 **自动配置**: 自动设置所需的环境变量
- 💻 **跨平台**: 支持 Windows, macOS 和 Linux

## 📋 支持的语言

| 语言 | 状态 | 特性 |
|------|------|------|
| Node.js | ✅ | 完整支持 |
| Go | ✅ | 完整支持 |
| Java | ✅ | 完整支持 |
| Python | ✅ | 完整支持 |
| Rust | 🔜 | 计划中 |
| Swift | 🔜 | 计划中 |
| Deno | 🔜 | 计划中 |
| PHP | 🔜 | 计划中 |
| Dart | 🔜 | 计划中 |
| Kotlin | 🔜 | 计划中 |

## 🚀 快速开始

### 安装

```bash
# 从源码构建
git clone https://github.com/Eusen/svm.git
cd svm
go build -o svm.exe

# 或者下载预编译的二进制文件
# (链接将在发布后提供)
```

### 基本用法

```bash
# 列出可用版本
svm node list
svm go list
svm java list
svm python list

# 列出所有版本（不过滤）
svm node list -a
svm go list --all

# 只显示已安装的版本
svm node list -i
svm go list --installed

# 安装指定版本
svm node install 16.20.2
svm go install 1.24.1
svm java install 17
svm python install 3.12.9

# 切换版本
svm node use 16.20.2
svm go use 1.24.1
svm java use 17
svm python use 3.12.9

# 删除版本
svm node remove 14.21.3
svm go remove 1.23.0
svm java remove 11
svm python remove 3.11.8

# 显示当前使用的版本
svm node current
svm go current
svm java current
svm python current

# 配置安装目录
svm config set-install-dir D:\SDKs
svm config get-install-dir
```

## 🔧 环境变量设置

SVM 会自动处理所需的环境变量设置：

- **Node.js**: 设置 PATH
- **Go**: 设置 GOROOT 和 PATH
- **Java**: 设置 JAVA_HOME 和 PATH
- **Python**: 设置 PYTHONHOME 和 PATH

## 🏗️ 项目结构

```
svm/
├── cmd/               # 命令行界面
├── internal/          # 内部实现
│   ├── config/        # 配置管理
│   ├── sdk/           # SDK 实现
│   └── utils/         # 工具函数
├── main.go            # 程序入口
└── README.md          # 项目文档
```

## 🤔 为什么选择 SVM?

- **统一体验**: 使用相同的命令管理所有语言环境
- **简单直观**: 无需记忆复杂的命令和选项
- **自动化**: 自动处理环境变量和路径设置
- **轻量级**: 单一可执行文件，无需复杂安装

## 🔮 未来计划

- [ ] 添加更多语言支持
- [ ] 支持在线更新版本列表
- [ ] 添加图形用户界面
- [ ] 支持通过配置文件设置项目级别的SDK版本

## 🤝 贡献

欢迎贡献代码、报告问题或提出新功能建议！我们特别希望听到你的使用体验和改进意见。

1. Fork 这个仓库
2. 创建你的特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交你的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启一个 Pull Request

## 📝 反馈与支持

遇到问题或有新想法？请在 GitHub Issues 中告诉我们！我们非常重视你的反馈，它将帮助我们不断改进这个工具。

## 📜 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件

---

<div align="center">
  <sub>用 ❤️ 和 AI 构建</sub>
</div> 