package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnvManager 用于管理环境变量的结构体
type EnvManager struct {
	Name            string            // SDK名称
	HomeVar         string            // HOME环境变量名称
	HomePath        string            // HOME环境变量值
	BinPath         string            // 可执行文件路径
	ExcludeKeywords []string          // 需要从PATH中排除的关键字
	ExtraVars       map[string]string // 额外的环境变量
}

// SetEnv 设置环境变量
func (e *EnvManager) SetEnv(version string) error {
	if runtime.GOOS == "windows" {
		return e.setWindowsEnv(version)
	}
	return e.setUnixEnv()
}

// setWindowsEnv 设置Windows环境变量
func (e *EnvManager) setWindowsEnv(version string) error {
	// 获取当前系统PATH
	getPathCmd := `[Environment]::GetEnvironmentVariable('Path', 'Machine')`
	output, err := RunCommand("powershell", "-Command", getPathCmd)
	if err != nil {
		return fmt.Errorf("获取系统PATH失败: %w", err)
	}

	// 移除旧的相关路径
	paths := strings.Split(output, ";")
	newPaths := make([]string, 0)
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		shouldKeep := true
		for _, keyword := range e.ExcludeKeywords {
			if strings.Contains(strings.ToLower(p), strings.ToLower(keyword)) {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			newPaths = append(newPaths, p)
		}
	}
	newPaths = append([]string{e.BinPath}, newPaths...)

	// 设置新的PATH
	newPath := strings.Join(newPaths, ";")

	// 创建PowerShell脚本
	script := fmt.Sprintf(`
		$setPathCmd = "[Environment]::SetEnvironmentVariable('Path', '%s', 'Machine')"
		Invoke-Expression $setPathCmd
		
		# 同时更新当前会话的PATH
		$env:Path = '%s'
		
		Write-Host '已切换到 %s %s'
		Write-Host '系统环境变量已更新，当前会话的PATH也已更新'
	`, newPath, newPath, e.Name, version)

	// 如果HomeVar和HomePath都不为空，添加设置HOME环境变量的命令
	if e.HomeVar != "" && e.HomePath != "" {
		homeScript := fmt.Sprintf(`
		$setHomeCmd = "[Environment]::SetEnvironmentVariable('%s', '%s', 'Machine')"
		Invoke-Expression $setHomeCmd
		
		# 同时更新当前会话的环境变量
		$env:%s = '%s'
		`, e.HomeVar, e.HomePath, e.HomeVar, e.HomePath)
		script = homeScript + script
	}

	// 处理额外的环境变量
	if e.ExtraVars != nil && len(e.ExtraVars) > 0 {
		extraVarsScript := ""
		for key, value := range e.ExtraVars {
			extraVarsScript += fmt.Sprintf(`
		$setExtraVarCmd = "[Environment]::SetEnvironmentVariable('%s', '%s', 'Machine')"
		Invoke-Expression $setExtraVarCmd
		
		# 同时更新当前会话的环境变量
		$env:%s = '%s'
		`, key, value, key, value)
		}
		script = extraVarsScript + script
	}

	// 将脚本保存到临时文件
	tempFile := filepath.Join(os.TempDir(), "svm_env.ps1")
	if err := os.WriteFile(tempFile, []byte(script), 0644); err != nil {
		return fmt.Errorf("创建临时脚本文件失败: %w", err)
	}
	defer os.Remove(tempFile)

	// 以管理员权限运行脚本
	adminCmd := fmt.Sprintf("Start-Process powershell -Verb RunAs -Wait -ArgumentList '-ExecutionPolicy Bypass -File %s'", tempFile)
	if _, err := RunCommand("powershell", "-Command", adminCmd); err != nil {
		return fmt.Errorf("设置环境变量失败: %w", err)
	}

	// 更新当前Go进程的环境变量
	os.Setenv("PATH", newPath)
	if e.HomeVar != "" && e.HomePath != "" {
		os.Setenv(e.HomeVar, e.HomePath)
	}

	// 更新当前进程的额外环境变量
	if e.ExtraVars != nil {
		for key, value := range e.ExtraVars {
			os.Setenv(key, value)
		}
	}

	return nil
}

// setUnixEnv 设置Unix环境变量
func (e *EnvManager) setUnixEnv() error {
	// 设置HOME环境变量
	if err := os.Setenv(e.HomeVar, e.HomePath); err != nil {
		return fmt.Errorf("设置%s失败: %w", e.HomeVar, err)
	}

	// 设置PATH
	path := os.Getenv("PATH")
	newPath := fmt.Sprintf("%s%s%s", e.BinPath, string(os.PathListSeparator), path)
	if err := os.Setenv("PATH", newPath); err != nil {
		return fmt.Errorf("设置PATH失败: %w", err)
	}

	// 设置额外的环境变量
	if e.ExtraVars != nil {
		for key, value := range e.ExtraVars {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("设置%s失败: %w", key, err)
			}
		}
	}

	return nil
}
