package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"svm/internal/config"
	"svm/internal/utils"

	"github.com/spf13/cobra"
)

func initGoCmd() {
	goCmd := &cobra.Command{
		Use:   "go",
		Short: "管理 Go 版本",
		Long:  `管理 Go 的不同版本，包括列出、安装、删除和切换版本。`,
	}

	goListCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有可用的 Go 版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			goSdk := GetSDK("go")

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

				// 获取Go安装目录
				installDir := filepath.Join(config.InstallDir, "go")

				// 检查目录是否存在
				if _, err := os.Stat(installDir); os.IsNotExist(err) {
					utils.Log.Info("未找到已安装的 Go 版本")
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
					utils.Log.Info("未找到已安装的 Go 版本")
					return nil
				}

				// 按版本号排序
				utils.SortVersionsDesc(installedVersions)

				// 获取当前使用的版本
				currentVersion, _ := goSdk.GetCurrentVersion()

				utils.Log.Info("已安装的 Go 版本：")
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
				versions, err = goSdk.ListAll()
			} else {
				// 显示过滤后的版本
				versions, err = goSdk.List()
			}

			if err != nil {
				return err
			}

			if all {
				utils.Log.Info("所有可用的 Go 版本：")
			} else {
				utils.Log.Info("可用的 Go 版本：")
			}

			for _, version := range versions {
				utils.Log.Custom(utils.IconStar, utils.Green, "", version)
			}
			return nil
		},
	}

	// 添加--installed或-i选项
	goListCmd.Flags().BoolP("installed", "i", false, "只显示已安装的版本")
	// 添加--all或-a选项
	goListCmd.Flags().BoolP("all", "a", false, "显示所有版本，不进行过滤")

	goInstallCmd := &cobra.Command{
		Use:   "install [version]",
		Short: "安装指定版本的 Go",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			goSdk := GetSDK("go")
			utils.Log.Install(fmt.Sprintf("正在安装 Go 版本 %s...", version))
			return goSdk.Install(version)
		},
	}

	goRemoveCmd := &cobra.Command{
		Use:   "remove [version]",
		Short: "删除指定版本的 Go",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			goSdk := GetSDK("go")
			utils.Log.Delete(fmt.Sprintf("正在删除 Go 版本 %s...", version))
			return goSdk.Remove(version)
		},
	}

	goUseCmd := &cobra.Command{
		Use:   "use [version]",
		Short: "切换到指定版本的 Go",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			goSdk := GetSDK("go")
			utils.Log.Switch(fmt.Sprintf("正在切换到 Go 版本 %s...", version))
			return goSdk.Use(version)
		},
	}

	goCurrentCmd := &cobra.Command{
		Use:   "current",
		Short: "显示当前使用的 Go 版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			goSdk := GetSDK("go")
			version, err := goSdk.GetCurrentVersion()
			if err != nil {
				utils.Log.Info("当前未设置 Go 版本")
				return nil
			}

			if version == "" {
				utils.Log.Info("当前未设置 Go 版本")
			} else {
				utils.Log.Info("当前使用的 Go 版本:")
				utils.Log.Custom(utils.IconHeart, utils.Magenta, "", version)
			}
			return nil
		},
	}

	goCmd.AddCommand(goListCmd, goInstallCmd, goRemoveCmd, goUseCmd, goCurrentCmd)
	rootCmd.AddCommand(goCmd)
}
