package cmd

import (
	"svm/internal/sdk"
	"svm/internal/utils"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "svm",
	Short: utils.FormatCommandTitle("SDK Version Manager - 管理多种编程语言的运行环境"),
	Long: utils.FormatCommandText(`SDK Version Manager (svm) 是一个命令行工具，用于管理多种编程语言的运行环境。
支持的语言包括：Node.js、Go、Python、Java、.NET 等。

使用示例：`) + `
  ` + utils.FormatCommandExample("svm node list") + `                列出所有可用的 Node.js 版本
  ` + utils.FormatCommandExample("svm java install 17") + `          安装 Java 17
  ` + utils.FormatCommandExample("svm node remove 10") + `           删除 Node.js 10
  ` + utils.FormatCommandExample("svm go use 1.24.1") + `            切换到 Go 1.24.1
  ` + utils.FormatCommandExample("svm dotnet sdk list") + `          列出所有可用的 .NET SDK 版本
  ` + utils.FormatCommandExample("svm dotnet asp-core install 7.0.0") + `  安装 ASP.NET Core 7.0.0 运行时`,
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
	registerSDK("dotnet", sdk.NewDotNetSDK())

	// 初始化各种命令
	initNodeCmd()
	initGoCmd()
	initJavaCmd()
	initPythonCmd()
	initDotNetCmd()
	initConfigCmd()

	// 为所有命令添加彩色输出
	formatCommandHelp(rootCmd)
}

// registerSDK 注册SDK实例
func registerSDK(name string, sdkInstance sdk.SDK) {
	sdkRegistry[name] = sdkInstance
}

// GetSDK 获取指定名称的SDK实例
func GetSDK(name string) sdk.SDK {
	return sdkRegistry[name]
}

// formatCommandHelp 为命令的帮助信息添加彩色输出
func formatCommandHelp(cmd *cobra.Command) {
	cmd.Short = utils.FormatCommandTitle(cmd.Short)
	if cmd.Long != "" {
		cmd.Long = utils.FormatCommandText(cmd.Long)
	}

	// 递归处理子命令
	for _, subCmd := range cmd.Commands() {
		formatCommandHelp(subCmd)
	}
}
