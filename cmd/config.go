package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"svm/internal/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理SVM配置",
	Long:  `管理SVM配置，包括安装目录等设置`,
}

var setInstallDirCmd = &cobra.Command{
	Use:   "set-install-dir <directory>",
	Short: "设置SDK安装目录",
	Long:  `设置SDK的安装目录，所有SDK将会安装到这个目录下`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取目标目录的绝对路径
		dir := args[0]
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("获取绝对路径失败: %w", err)
		}

		// 创建目录（如果不存在）
		if err := os.MkdirAll(absDir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}

		// 加载配置
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		// 如果目录已经存在且不为空，提示用户
		entries, err := os.ReadDir(absDir)
		if err != nil {
			return fmt.Errorf("读取目录失败: %w", err)
		}
		if len(entries) > 0 {
			fmt.Println("警告：目标目录不为空，现有的SDK将保持在原位置")
		}

		// 更新配置
		if err := cfg.SetInstallDir(absDir); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Printf("已将安装目录设置为: %s\n", absDir)
		return nil
	},
}

var getInstallDirCmd = &cobra.Command{
	Use:   "get-install-dir",
	Short: "获取当前的SDK安装目录",
	Long:  `显示当前配置的SDK安装目录`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("加载配置失败: %w", err)
		}

		fmt.Printf("当前安装目录: %s\n", cfg.InstallDir)
		return nil
	},
}

func initConfigCmd() {
	configCmd.AddCommand(setInstallDirCmd)
	configCmd.AddCommand(getInstallDirCmd)
	rootCmd.AddCommand(configCmd)
} 