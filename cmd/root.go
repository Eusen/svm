package cmd

import (
	"svm/internal/sdk"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "svm",
	Short: "SDK Version Manager - 管理多种编程语言的运行环境",
	Long: `SDK Version Manager (svm) 是一个命令行工具，用于管理多种编程语言的运行环境。
支持的语言包括：Node.js、Go、Python、Java 等。

使用示例：
  svm node list          列出所有可用的 Node.js 版本
  svm java install 17    安装 Java 17
  svm node remove 10     删除 Node.js 10
  svm go use 1.24.1      切换到 Go 1.24.1`,
}

// 全局SDK实例的映射
var sdkRegistry = map[string]sdk.SDK{}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 初始化所有SDK
	registerSDK("node", sdk.NewNodeSDK())
	registerSDK("go", sdk.NewGoSDK())
	registerSDK("java", sdk.NewJavaSDK())
	registerSDK("python", sdk.NewPythonSDK())

	// 初始化各种命令
	initNodeCmd()
	initGoCmd()
	initJavaCmd()
	initPythonCmd()
	initConfigCmd()
}

// registerSDK 注册SDK实例
func registerSDK(name string, sdkInstance sdk.SDK) {
	sdkRegistry[name] = sdkInstance
}

// GetSDK 获取指定名称的SDK实例
func GetSDK(name string) sdk.SDK {
	return sdkRegistry[name]
}
