package sdk

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"svm/internal/config"
	"svm/internal/utils"
)

// PythonVersion 表示Python版本信息
type PythonVersion struct {
	Version     string `json:"version"`
	ReleaseURL  string `json:"release_url"`
	ReleaseDate string `json:"release_date"`
}

// PythonSDKProvider 实现了SDKProvider接口
type PythonSDKProvider struct {
	config *config.Config
}

// NewPythonSDK 创建一个新的Python SDK
func NewPythonSDK() SDK {
	provider := &PythonSDKProvider{
		config: nil, // 这里为空，会由BaseSDK初始化时设置
	}

	return &pythonSDK{
		BaseSDK: *NewBaseSDK("python", provider, DefaultVersionPrefixHandlers()),
	}
}

// pythonSDK 是Python SDK的具体实现
type pythonSDK struct {
	BaseSDK
}

// GetVersionList 实现SDKProvider接口，获取所有可用的Python版本
func (p *PythonSDKProvider) GetVersionList() ([]string, error) {
	// 直接从Python官方FTP目录获取版本列表
	ftpUrl := "https://www.python.org/ftp/python/"

	// 获取目录列表
	resp, err := http.Get(ftpUrl)
	if err != nil {
		return nil, fmt.Errorf("获取Python版本列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取Python版本列表失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %w", err)
	}

	// 解析HTML内容，提取版本目录
	bodyStr := string(body)

	// 使用正则表达式匹配版本目录
	// 匹配形如 href="3.10.0/" 的目录链接
	versionRegex := regexp.MustCompile(`href="(\d+\.\d+\.\d+)/"`)
	matches := versionRegex.FindAllStringSubmatch(bodyStr, -1)

	// 如果上面的正则表达式没有匹配到，尝试另一种格式
	if len(matches) == 0 {
		versionRegex = regexp.MustCompile(`>(\d+\.\d+\.\d+)/<`)
		matches = versionRegex.FindAllStringSubmatch(bodyStr, -1)
	}

	// 按次版本分组，每个次版本只保留最新的版本
	minorVersions := make(map[string]string)

	for _, match := range matches {
		if len(match) > 1 {
			version := match[1]

			// 提取次版本号（例如：3.10.0 -> 3.10）
			parts := strings.Split(version, ".")
			if len(parts) >= 2 {
				minorVersion := parts[0] + "." + parts[1]

				// 如果该次版本尚未记录或当前版本更新，则更新
				if currentVersion, exists := minorVersions[minorVersion]; !exists {
					minorVersions[minorVersion] = version
				} else {
					// 使用utils.compareVersions函数进行版本比较
					if utils.CompareVersionsStr(version, currentVersion) > 0 {
						minorVersions[minorVersion] = version
					}
				}
			} else {
				// 对于格式不符合预期的版本，直接添加
				minorVersions[version] = version
			}
		}
	}

	// 如果没有找到版本，返回错误
	if len(minorVersions) == 0 {
		return nil, fmt.Errorf("未找到Python版本")
	}

	// 将次版本映射转换为版本列表
	var versionList []string
	for _, version := range minorVersions {
		versionList = append(versionList, version)
	}

	// 按版本号排序（从新到旧）
	utils.SortVersionsDesc(versionList)

	return versionList, nil
}

// GetAllVersionList 实现SDKProvider接口，获取所有可用的Python版本（不过滤）
func (p *PythonSDKProvider) GetAllVersionList() ([]string, error) {
	// 直接从Python官方FTP目录获取版本列表
	ftpUrl := "https://www.python.org/ftp/python/"

	// 获取目录列表
	resp, err := http.Get(ftpUrl)
	if err != nil {
		return nil, fmt.Errorf("获取Python版本列表失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取Python版本列表失败，HTTP状态码: %d", resp.StatusCode)
	}

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应内容失败: %w", err)
	}

	// 解析HTML内容，提取版本目录
	bodyStr := string(body)

	// 使用正则表达式匹配版本目录
	// 匹配形如 href="3.10.0/" 的目录链接
	versionRegex := regexp.MustCompile(`href="(\d+\.\d+\.\d+)/"`)
	matches := versionRegex.FindAllStringSubmatch(bodyStr, -1)

	// 如果上面的正则表达式没有匹配到，尝试另一种格式
	if len(matches) == 0 {
		versionRegex = regexp.MustCompile(`>(\d+\.\d+\.\d+)/<`)
		matches = versionRegex.FindAllStringSubmatch(bodyStr, -1)
	}

	// 提取所有版本号
	var versionList []string
	for _, match := range matches {
		if len(match) > 1 {
			versionList = append(versionList, match[1])
		}
	}

	// 如果没有找到版本，返回错误
	if len(versionList) == 0 {
		return nil, fmt.Errorf("未找到Python版本")
	}

	// 按版本号排序（从新到旧）
	utils.SortVersionsDesc(versionList)

	return versionList, nil
}

// GetDownloadURL 构建Python下载URL
func (p *PythonSDKProvider) GetDownloadURL(version, osName, arch string) string {
	// 根据操作系统和架构构建下载URL
	baseUrl := "https://www.python.org/ftp/python"

	// 尝试不同的下载格式
	if osName == "windows" {
		// 确定架构
		archSuffix := ""
		if arch == "x64" || arch == "amd64" {
			archSuffix = "-amd64"
		} else if arch == "arm64" {
			archSuffix = "-arm64"
		}

		// 构建基本URL路径
		basePath := fmt.Sprintf("%s/%s", baseUrl, version)

		// 尝试获取目录列表
		resp, err := http.Get(basePath + "/")
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				bodyStr := string(body)

				// 查找所有zip文件链接
				zipRegex := regexp.MustCompile(`href="([^"]+\.zip)"`)
				matches := zipRegex.FindAllStringSubmatch(bodyStr, -1)

				// 优先选择非嵌入式版本
				var regularZip string
				var embedZip string

				for _, match := range matches {
					if len(match) > 1 {
						fileName := match[1]

						// 检查是否包含架构后缀
						if strings.Contains(fileName, archSuffix) {
							fullUrl := fmt.Sprintf("%s/%s", basePath, fileName)

							// 检查是否是嵌入式版本
							if strings.Contains(fileName, "embed") {
								embedZip = fullUrl
							} else {
								regularZip = fullUrl
								break // 找到非嵌入式版本就停止
							}
						}
					}
				}

				// 优先返回非嵌入式版本，如果没有则返回嵌入式版本
				if regularZip != "" {
					return regularZip
				}

				if embedZip != "" {
					return embedZip
				}
			}
		}

		// 如果无法获取目录列表或没有找到匹配的文件，使用默认URL
		// 优先尝试完整版本的ZIP包，这样会包含pip
		regularUrl := fmt.Sprintf("%s/%s/python-%s%s.zip", baseUrl, version, version, archSuffix)

		// 检查常规URL是否存在
		exists, _ := utils.CheckURLExists(regularUrl)
		if exists {
			return regularUrl
		}

		// 如果常规格式不存在，尝试嵌入式格式
		embedUrl := fmt.Sprintf("%s/%s/python-%s-embed%s.zip", baseUrl, version, version, archSuffix)
		exists, _ = utils.CheckURLExists(embedUrl)
		if exists {
			return embedUrl
		}

		// 如果两种格式都不存在，返回常规格式，让BaseSDK处理失败情况
		return regularUrl
	} else if osName == "darwin" {
		// macOS使用pkg安装包
		if arch == "arm64" {
			return fmt.Sprintf("%s/%s/python-%s-macos11.pkg", baseUrl, version, version)
		}
		return fmt.Sprintf("%s/%s/python-%s-macosx10.9.pkg", baseUrl, version, version)
	} else {
		// Linux通常使用源码包
		return fmt.Sprintf("%s/%s/Python-%s.tgz", baseUrl, version, version)
	}
}

// GetExtractDir 获取解压后的目录名
func (p *PythonSDKProvider) GetExtractDir(version, downloadedFile string) string {
	if runtime.GOOS == "linux" {
		return fmt.Sprintf("Python-%s", version)
	}
	return "" // Windows和macOS不需要特殊处理
}

// GetBinDir 获取bin目录
func (p *PythonSDKProvider) GetBinDir(baseDir string) string {
	scriptsDir := filepath.Join(baseDir, "Scripts")

	// 检查Scripts目录是否存在
	if _, err := os.Stat(scriptsDir); err == nil {
		return fmt.Sprintf("%s;%s", baseDir, scriptsDir)
	}

	// 如果Scripts目录不存在，只返回baseDir
	return baseDir
}

// ConfigureEnv 配置环境变量
func (p *PythonSDKProvider) ConfigureEnv(version, installDir string) ([]config.EnvVar, error) {
	// 添加Python主目录和Scripts目录到PATH
	scriptsDir := filepath.Join(installDir, "Scripts")

	// 确保目录存在
	if _, err := os.Stat(installDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Python安装目录不存在: %s", installDir)
	}

	// 检查Scripts目录是否存在，如果不存在则不添加到PATH
	scriptsPath := installDir
	if _, err := os.Stat(scriptsDir); err == nil {
		scriptsPath = fmt.Sprintf("%s;%s", installDir, scriptsDir)
	}

	// 创建排除关键字列表
	excludeKeywords := []string{"python"}

	return []config.EnvVar{
		{
			Key:   "PYTHONHOME",
			Value: installDir,
		},
		{
			Key:   "PATH",
			Value: scriptsPath,
		},
		{
			Key:   "EXCLUDE_KEYWORDS",
			Value: strings.Join(excludeKeywords, ","),
		},
	}, nil
}

// PreInstall 安装前的准备工作
func (p *PythonSDKProvider) PreInstall(version string) error {
	// 对于Python，不需要特殊的安装前准备
	return nil
}

// PostInstall 安装后的处理工作
func (p *PythonSDKProvider) PostInstall(version, installDir string) error {
	// 检查是否是嵌入式Python包
	isEmbedded := false
	entries, err := os.ReadDir(installDir)
	if err != nil {
		return fmt.Errorf("读取安装目录失败: %w", err)
	}

	// 检查是否包含python.exe和python*._pth文件，这是嵌入式Python的特征
	var pthFile string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, "._pth") && strings.Contains(name, "python") {
			isEmbedded = true
			pthFile = filepath.Join(installDir, name)
			break
		}
	}

	// 如果是嵌入式Python，尝试添加pip支持
	if isEmbedded {
		fmt.Println("检测到嵌入式Python包，尝试添加pip支持...")

		// 1. 修改._pth文件，取消注释import site
		if pthFile != "" {
			content, err := os.ReadFile(pthFile)
			if err != nil {
				fmt.Printf("读取._pth文件失败: %v\n", err)
			} else {
				// 取消注释import site
				newContent := strings.Replace(string(content), "#import site", "import site", 1)
				if err := os.WriteFile(pthFile, []byte(newContent), 0644); err != nil {
					fmt.Printf("修改._pth文件失败: %v\n", err)
				} else {
					fmt.Println("成功修改._pth文件，启用site模块")
				}
			}
		}

		// 2. 下载get-pip.py
		getPipURL := "https://bootstrap.pypa.io/get-pip.py"
		getPipPath := filepath.Join(installDir, "get-pip.py")

		fmt.Println("下载get-pip.py...")
		resp, err := http.Get(getPipURL)
		if err != nil {
			fmt.Printf("下载get-pip.py失败: %v\n", err)
		} else {
			defer resp.Body.Close()

			// 保存get-pip.py
			out, err := os.Create(getPipPath)
			if err != nil {
				fmt.Printf("创建get-pip.py文件失败: %v\n", err)
			} else {
				defer out.Close()

				_, err = io.Copy(out, resp.Body)
				if err != nil {
					fmt.Printf("保存get-pip.py失败: %v\n", err)
				} else {
					// 3. 运行get-pip.py
					fmt.Println("运行get-pip.py安装pip...")
					pythonExe := filepath.Join(installDir, "python.exe")
					cmd := exec.Command(pythonExe, getPipPath, "--no-warn-script-location")
					output, err := cmd.CombinedOutput()
					if err != nil {
						fmt.Printf("安装pip失败: %v\n%s\n", err, string(output))
					} else {
						fmt.Println("pip安装成功")

						// 4. 创建Scripts目录（如果不存在）
						scriptsDir := filepath.Join(installDir, "Scripts")
						if err := os.MkdirAll(scriptsDir, 0755); err != nil {
							fmt.Printf("创建Scripts目录失败: %v\n", err)
						}
					}
				}
			}
		}

		return nil
	}

	// 检查是否是完整版本的Python（包含Lib目录和Scripts目录）
	hasLib := false
	hasScripts := false
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() && name == "Lib" {
			hasLib = true
		}
		if entry.IsDir() && name == "Scripts" {
			hasScripts = true
		}
	}

	// 如果是完整版本的Python，检查是否有pip
	if hasLib && hasScripts {
		fmt.Println("检测到完整版本的Python")

		// 检查pip是否存在
		pipPath := filepath.Join(installDir, "Scripts", "pip.exe")
		if _, err := os.Stat(pipPath); os.IsNotExist(err) {
			// 如果pip不存在，尝试安装
			fmt.Println("未检测到pip，尝试安装...")

			// 使用ensurepip模块安装pip
			pythonExe := filepath.Join(installDir, "python.exe")
			cmd := exec.Command(pythonExe, "-m", "ensurepip", "--upgrade")
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("安装pip失败: %v\n%s\n", err, string(output))
				// 继续执行，不返回错误
			} else {
				fmt.Println("pip安装成功")
			}
		} else {
			fmt.Println("检测到pip已安装")
		}

		return nil
	}

	// 对于Windows，我们需要运行安装程序
	if runtime.GOOS == "windows" {
		// 查找安装程序
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasSuffix(name, ".exe") && strings.Contains(name, "python") {
				// 发现安装程序
				installer := filepath.Join(installDir, name)

				// 检查是否是Python可执行文件而不是安装程序
				if strings.Contains(name, "embed") || name == "python.exe" {
					fmt.Println("检测到Python可执行文件，无需额外安装")
					return nil
				}

				// 创建安装目标目录
				targetDir := installDir
				// 不再创建额外的子目录，直接使用installDir

				// 静默安装
				fmt.Printf("正在安装Python到 %s\n", targetDir)
				cmd := exec.Command(installer, "/quiet", "InstallAllUsers=0",
					fmt.Sprintf("TargetDir=%s", targetDir),
					"Include_test=0", "Include_tools=1", "PrependPath=1")

				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("安装Python失败: %w\n%s", err, string(output))
				}

				fmt.Printf("Python安装完成: %s\n", targetDir)
				break
			}
		}
	} else if runtime.GOOS == "darwin" {
		// 对于macOS，我们需要挂载和安装pkg
		entries, err := os.ReadDir(installDir)
		if err != nil {
			return fmt.Errorf("读取安装目录失败: %w", err)
		}

		for _, entry := range entries {
			name := entry.Name()
			if strings.HasSuffix(name, ".pkg") && strings.Contains(name, "python") {
				// 发现安装包
				pkg := filepath.Join(installDir, name)

				// 安装pkg
				installCmd := fmt.Sprintf(`installer -pkg "%s" -target CurrentUserHomeDirectory`, pkg)
				fmt.Printf("正在安装Python: %s\n", installCmd)
				cmd := exec.Command("bash", "-c", installCmd)
				output, err := cmd.CombinedOutput()
				if err != nil {
					return fmt.Errorf("安装Python失败: %w\n%s", err, string(output))
				}

				fmt.Printf("Python安装完成\n")
				break
			}
		}
	} else {
		// 对于Linux，我们需要编译源码
		extractDir := filepath.Join(installDir, fmt.Sprintf("Python-%s", version))
		if _, err := os.Stat(extractDir); err == nil {
			// 编译源码
			configureCmd := fmt.Sprintf(`cd "%s" && ./configure --prefix="%s" && make && make install`,
				extractDir, installDir)
			fmt.Printf("正在编译Python: %s\n", configureCmd)
			cmd := exec.Command("bash", "-c", configureCmd)
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("编译Python失败: %w\n%s", err, string(output))
			}

			fmt.Printf("Python编译和安装完成\n")

			// 删除源码目录
			if err := os.RemoveAll(extractDir); err != nil {
				fmt.Printf("警告：删除源码目录失败: %v\n", err)
			}
		}
	}

	return nil
}

// GetArchiveType 获取归档类型
func (p *PythonSDKProvider) GetArchiveType() string {
	// 对于Python，我们根据操作系统和具体的下载文件来决定归档类型
	// 这个方法将在实际下载文件后由调用方通过 GetArchiveTypeForFile 获取具体类型
	return "auto"
}

// GetArchiveTypeForFile 根据文件名确定正确的归档类型
func (p *PythonSDKProvider) GetArchiveTypeForFile(filePath string) string {
	// 根据文件扩展名判断
	fileName := filepath.Base(filePath)
	if strings.HasSuffix(fileName, ".zip") {
		return "zip"
	} else if strings.HasSuffix(fileName, ".tar.gz") || strings.HasSuffix(fileName, ".tgz") {
		return "tar.gz"
	} else if strings.HasSuffix(fileName, ".exe") || strings.HasSuffix(fileName, ".msi") {
		// 可执行安装程序
		return "none"
	}

	// 默认情况下尝试作为zip处理
	return "zip"
}
