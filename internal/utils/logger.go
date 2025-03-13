package utils

import (
	"fmt"
	"runtime"
	"strings"
)

// 定义颜色常量
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
)

// 定义图标常量
const (
	IconInfo     = "ℹ️"
	IconSuccess  = "✅"
	IconWarning  = "⚠️"
	IconError    = "❌"
	IconInstall  = "📦"
	IconDownload = "📥"
	IconExtract  = "📂"
	IconConfig   = "⚙️"
	IconSwitch   = "🔄"
	IconSearch   = "🔍"
	IconMove     = "🚚"
	IconLink     = "🔗"
	IconDelete   = "🗑️"
	IconCheck    = "✓"
	IconHeart    = "💖"
	IconStar     = "⭐"
)

// Logger 提供美化的日志输出功能
type Logger struct {
	useColors bool
	useIcons  bool
}

// NewLogger 创建一个新的 Logger 实例
func NewLogger() *Logger {
	// Windows 命令行默认不支持 ANSI 颜色，但 Windows 10+ 的新终端支持
	useColors := runtime.GOOS != "windows" || isWindowsTerminalSupported()

	return &Logger{
		useColors: useColors,
		useIcons:  true,
	}
}

// isWindowsTerminalSupported 检查 Windows 终端是否支持 ANSI 颜色
func isWindowsTerminalSupported() bool {
	// 简单实现，实际上可能需要更复杂的检测
	return true
}

// formatMessage 格式化消息，添加颜色和图标
func (l *Logger) formatMessage(icon, color, prefix, message string) string {
	var result strings.Builder

	if l.useIcons {
		result.WriteString(icon + " ")
	}

	if l.useColors {
		result.WriteString(color)
	}

	if prefix != "" {
		result.WriteString(prefix + ": ")
	}

	result.WriteString(message)

	if l.useColors {
		result.WriteString(Reset)
	}

	return result.String()
}

// Info 输出信息级别的日志
func (l *Logger) Info(message string) {
	fmt.Println(l.formatMessage(IconInfo, Cyan, "INFO", message))
}

// Success 输出成功级别的日志
func (l *Logger) Success(message string) {
	fmt.Println(l.formatMessage(IconSuccess, Green, "成功", message))
}

// Warning 输出警告级别的日志
func (l *Logger) Warning(message string) {
	fmt.Println(l.formatMessage(IconWarning, Yellow, "警告", message))
}

// Error 输出错误级别的日志
func (l *Logger) Error(message string) {
	fmt.Println(l.formatMessage(IconError, Red, "错误", message))
}

// Install 输出安装相关的日志
func (l *Logger) Install(message string) {
	fmt.Println(l.formatMessage(IconInstall, Magenta, "安装", message))
}

// Download 输出下载相关的日志
func (l *Logger) Download(message string) {
	fmt.Println(l.formatMessage(IconDownload, Blue, "下载", message))
}

// Extract 输出解压相关的日志
func (l *Logger) Extract(message string) {
	fmt.Println(l.formatMessage(IconExtract, Yellow, "解压", message))
}

// Config 输出配置相关的日志
func (l *Logger) Config(message string) {
	fmt.Println(l.formatMessage(IconConfig, Cyan, "配置", message))
}

// Switch 输出切换版本相关的日志
func (l *Logger) Switch(message string) {
	fmt.Println(l.formatMessage(IconSwitch, Green, "切换", message))
}

// Move 输出移动文件相关的日志
func (l *Logger) Move(message string) {
	fmt.Println(l.formatMessage(IconMove, Yellow, "移动", message))
}

// Link 输出创建链接相关的日志
func (l *Logger) Link(message string) {
	fmt.Println(l.formatMessage(IconLink, Cyan, "链接", message))
}

// Delete 输出删除文件相关的日志
func (l *Logger) Delete(message string) {
	fmt.Println(l.formatMessage(IconDelete, Red, "删除", message))
}

// Check 输出检查相关的日志
func (l *Logger) Check(message string) {
	fmt.Println(l.formatMessage(IconCheck, Green, "检查", message))
}

// Custom 输出自定义图标和颜色的日志
func (l *Logger) Custom(icon, color, prefix, message string) {
	fmt.Println(l.formatMessage(icon, color, prefix, message))
}

// DisableColors 禁用颜色输出
func (l *Logger) DisableColors() {
	l.useColors = false
}

// EnableColors 启用颜色输出
func (l *Logger) EnableColors() {
	l.useColors = true
}

// DisableIcons 禁用图标输出
func (l *Logger) DisableIcons() {
	l.useIcons = false
}

// EnableIcons 启用图标输出
func (l *Logger) EnableIcons() {
	l.useIcons = true
}

// Search 输出搜索相关的日志
func (l *Logger) Search(message string) {
	fmt.Println(l.formatMessage(IconSearch, Blue, "搜索", message))
}

// 全局 Logger 实例
var Log = NewLogger()

// FormatCommandText 格式化命令的描述文本，添加颜色
func FormatCommandText(text string) string {
	if Log.useColors {
		return fmt.Sprintf("%s%s%s", Cyan, text, Reset)
	}
	return text
}

// FormatCommandTitle 格式化命令的标题，添加颜色
func FormatCommandTitle(text string) string {
	if Log.useColors {
		return fmt.Sprintf("%s%s%s%s", Bold, Green, text, Reset)
	}
	return text
}

// FormatCommandExample 格式化命令的示例，添加颜色
func FormatCommandExample(text string) string {
	if Log.useColors {
		return fmt.Sprintf("%s%s%s", Yellow, text, Reset)
	}
	return text
}
