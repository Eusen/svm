package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"svm/internal/config"
	"svm/internal/utils"

	"github.com/spf13/cobra"
)

func initJavaCmd() {
	javaCmd := &cobra.Command{
		Use:   "java",
		Short: "管理 Java 版本",
		Long:  `管理 Java 的不同版本，包括列出、安装、删除和切换版本。`,
	}

	javaListCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有可用的 Java 版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			javaSdk := GetSDK("java")

			// 检查是否只显示已安装的版本
			installed, _ := cmd.Flags().GetBool("installed")
			// 检查是否显示所有版本
			all, _ := cmd.Flags().GetBool("all")

			if installed {
				// 获取安装目录
				config, err := config.LoadConfig()
				if err != nil {
					return err
				}

				// 获取Java安装目录
				installDir := filepath.Join(config.InstallDir, "java")

				// 检查目录是否存在
				if _, err := os.Stat(installDir); os.IsNotExist(err) {
					utils.Log.Info("未找到已安装的 Java 版本")
					return nil
				}

				// 读取安装目录中的所有子目录
				entries, err := os.ReadDir(installDir)
				if err != nil {
					return err
				}

				// 过滤出版本目录
				var installedVersions []string
				for _, entry := range entries {
					if entry.IsDir() {
						installedVersions = append(installedVersions, entry.Name())
					}
				}

				if len(installedVersions) == 0 {
					utils.Log.Info("未找到已安装的 Java 版本")
					return nil
				}

				// 按版本号排序
				utils.SortVersionsDesc(installedVersions)

				// 获取当前使用的版本
				currentVersion, _ := javaSdk.GetCurrentVersion()

				utils.Log.Info("已安装的 Java 版本：")
				for _, version := range installedVersions {
					if version == currentVersion {
						utils.Log.Custom(utils.IconHeart, utils.Magenta, "", version+" (当前使用)")
					} else {
						utils.Log.Custom(utils.IconStar, utils.Green, "", version)
					}
				}
				return nil
			}

			// 显示可用版本
			var versions []string
			var err error

			if all {
				// 显示所有版本
				versions, err = javaSdk.ListAll()
			} else {
				// 显示过滤后的版本
				versions, err = javaSdk.List()
			}

			if err != nil {
				return err
			}

			if all {
				utils.Log.Info("所有可用的 Java 版本：")
			} else {
				utils.Log.Info("可用的 Java 版本：")
			}

			for _, version := range versions {
				utils.Log.Custom(utils.IconStar, utils.Green, "", version)
			}
			return nil
		},
	}

	// 添加--installed或-i选项
	javaListCmd.Flags().BoolP("installed", "i", false, "只显示已安装的版本")
	// 添加--all或-a选项
	javaListCmd.Flags().BoolP("all", "a", false, "显示所有版本，不进行过滤")

	javaInstallCmd := &cobra.Command{
		Use:   "install [version]",
		Short: "安装指定版本的 Java",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			javaSdk := GetSDK("java")
			utils.Log.Install(fmt.Sprintf("正在安装 Java 版本 %s...", version))
			return javaSdk.Install(version)
		},
	}

	javaRemoveCmd := &cobra.Command{
		Use:   "remove [version]",
		Short: "删除指定版本的 Java",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			javaSdk := GetSDK("java")
			utils.Log.Delete(fmt.Sprintf("正在删除 Java 版本 %s...", version))
			return javaSdk.Remove(version)
		},
	}

	javaUseCmd := &cobra.Command{
		Use:   "use [version]",
		Short: "切换到指定版本的 Java",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			javaSdk := GetSDK("java")
			utils.Log.Switch(fmt.Sprintf("正在切换到 Java 版本 %s...", version))
			return javaSdk.Use(version)
		},
	}

	javaCurrentCmd := &cobra.Command{
		Use:   "current",
		Short: "显示当前使用的 Java 版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			javaSdk := GetSDK("java")
			version, err := javaSdk.GetCurrentVersion()
			if err != nil {
				// 不返回错误，而是显示友好的消息
				utils.Log.Info("当前未设置 Java 版本")
				return nil
			}

			if version == "" {
				utils.Log.Info("当前未设置 Java 版本")
			} else {
				utils.Log.Info("当前使用的 Java 版本:")
				utils.Log.Custom(utils.IconHeart, utils.Magenta, "", version)
			}
			return nil
		},
	}

	javaCmd.AddCommand(javaListCmd, javaInstallCmd, javaRemoveCmd, javaUseCmd, javaCurrentCmd)
	rootCmd.AddCommand(javaCmd)
}
