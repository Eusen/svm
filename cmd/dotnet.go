package cmd

import (
	"fmt"
	"svm/internal/utils"

	"github.com/spf13/cobra"
)

func initDotNetCmd() {
	dotnetCmd := &cobra.Command{
		Use:   "dotnet",
		Short: "管理 .NET 版本",
		Long:  `管理 .NET 的不同版本，包括SDK和各种运行时。`,
	}

	// 为每种组件类型创建子命令
	sdkCmd := createComponentCmd("sdk", "SDK（软件开发工具包）")
	aspCoreCmd := createComponentCmd("asp-core", "ASP.NET Core 运行时")
	desktopCmd := createComponentCmd("desktop", "桌面运行时")
	runtimeCmd := createComponentCmd("runtime", ".NET 运行时")

	// 添加子命令
	dotnetCmd.AddCommand(sdkCmd, aspCoreCmd, desktopCmd, runtimeCmd)

	rootCmd.AddCommand(dotnetCmd)
}

// 创建组件子命令
func createComponentCmd(componentType, description string) *cobra.Command {
	componentCmd := &cobra.Command{
		Use:   componentType,
		Short: "管理 .NET " + description,
		Long:  `管理 .NET ` + description + ` 的不同版本，包括列出、安装、删除和切换版本。`,
	}

	// 添加操作子命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有可用的 .NET " + description + " 版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDotNetCommand("list", componentType, args)
		},
	}

	// 添加标志
	listCmd.Flags().BoolP("installed", "i", false, "只显示已安装的版本")
	listCmd.Flags().BoolP("all", "a", false, "显示所有版本（不过滤）")

	installCmd := &cobra.Command{
		Use:   "install <version>",
		Short: "安装指定版本的 .NET " + description,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDotNetCommand("install", componentType, args)
		},
	}

	useCmd := &cobra.Command{
		Use:   "use <version>",
		Short: "切换到指定版本的 .NET " + description,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDotNetCommand("use", componentType, args)
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove <version>",
		Short: "删除指定版本的 .NET " + description,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDotNetCommand("remove", componentType, args)
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "显示当前使用的 .NET " + description + " 版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDotNetCommand("current", componentType, args)
		},
	}

	componentCmd.AddCommand(listCmd, installCmd, useCmd, removeCmd, currentCmd)

	return componentCmd
}

// 处理.NET命令
func handleDotNetCommand(action, componentType string, args []string) error {
	// 获取SDK实例
	dotnetSdk := GetSDK("dotnet")

	// 设置组件类型
	if setter, ok := dotnetSdk.(interface{ SetComponentType(string) }); ok {
		setter.SetComponentType(componentType)
	} else {
		return fmt.Errorf("无法设置 .NET 组件类型")
	}

	// 根据操作执行相应的功能
	switch action {
	case "list":
		versions, err := dotnetSdk.List()
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			utils.Log.Info(fmt.Sprintf("未找到可用的 .NET %s 版本", getComponentTypeDescription(componentType)))
			return nil
		}

		utils.Log.Info(fmt.Sprintf("可用的 .NET %s 版本:", getComponentTypeDescription(componentType)))
		for _, version := range versions {
			utils.Log.Custom(utils.IconStar, utils.Green, "", version)
		}

	case "install":
		version := args[0]
		utils.Log.Install(fmt.Sprintf("正在安装 .NET %s 版本 %s...", getComponentTypeDescription(componentType), version))
		return dotnetSdk.Install(version)

	case "use":
		version := args[0]
		return dotnetSdk.Use(version)

	case "remove":
		version := args[0]
		utils.Log.Delete(fmt.Sprintf("正在删除 .NET %s 版本 %s...", getComponentTypeDescription(componentType), version))
		return dotnetSdk.Remove(version)

	case "current":
		version, err := dotnetSdk.GetCurrentVersion()
		if err != nil {
			return err
		}

		if version == "" {
			utils.Log.Info(fmt.Sprintf("未设置当前 .NET %s 版本", getComponentTypeDescription(componentType)))
			return nil
		}

		utils.Log.Info(fmt.Sprintf("当前使用的 .NET %s 版本:", getComponentTypeDescription(componentType)))
		utils.Log.Custom(utils.IconHeart, utils.Magenta, "", version)
	}

	return nil
}

// 获取组件类型描述
func getComponentTypeDescription(componentType string) string {
	switch componentType {
	case "sdk":
		return "SDK"
	case "asp-core":
		return "ASP.NET Core 运行时"
	case "desktop":
		return "桌面运行时"
	case "runtime":
		return ".NET 运行时"
	default:
		return componentType
	}
}
