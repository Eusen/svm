package utils

import (
	"fmt"
	"net/http"
	"strings"
)

// FindBestMatchingVersion 查找最匹配的版本
// 如果找不到精确匹配的版本，会查找小于等于请求版本的最新版本
func FindBestMatchingVersion(requestedVersion string, availableVersions []string, stripPrefix string) (string, bool) {
	if len(availableVersions) == 0 {
		return "", false
	}

	// 标准化请求版本（去掉前缀）
	requestedVersionNormalized := requestedVersion
	if stripPrefix != "" && strings.HasPrefix(requestedVersionNormalized, stripPrefix) {
		requestedVersionNormalized = strings.TrimPrefix(requestedVersionNormalized, stripPrefix)
	}

	// 1. 查找精确匹配
	for _, v := range availableVersions {
		normalizedV := v
		if stripPrefix != "" && strings.HasPrefix(normalizedV, stripPrefix) {
			normalizedV = strings.TrimPrefix(normalizedV, stripPrefix)
		}

		if normalizedV == requestedVersionNormalized {
			return v, true
		}
	}

	// 特殊处理: 对于短版本号(如"8"，"11"等)，尝试前缀匹配
	if len(requestedVersionNormalized) <= 2 {
		for _, v := range availableVersions {
			normalizedV := v
			if stripPrefix != "" && strings.HasPrefix(normalizedV, stripPrefix) {
				normalizedV = strings.TrimPrefix(normalizedV, stripPrefix)
			}

			if strings.HasPrefix(normalizedV, requestedVersionNormalized+".") || normalizedV == requestedVersionNormalized {
				return v, true
			}
		}
	}

	// 2. 处理版本号中可能包含的'v'前缀问题
	if strings.HasPrefix(requestedVersionNormalized, "v") {
		noVPrefix := strings.TrimPrefix(requestedVersionNormalized, "v")
		for _, v := range availableVersions {
			normalizedV := v
			if stripPrefix != "" && strings.HasPrefix(normalizedV, stripPrefix) {
				normalizedV = strings.TrimPrefix(normalizedV, stripPrefix)
			}

			normalizedV = strings.TrimPrefix(normalizedV, "v")
			if normalizedV == noVPrefix {
				return v, true
			}
		}
	}

	// 3. 查找最接近但不超过请求版本的版本
	var parsedRequestVersion []int
	var err error

	// 尝试不同格式解析请求版本
	versionToParse := requestedVersionNormalized

	// 如果有v前缀，去掉
	if strings.HasPrefix(versionToParse, "v") {
		versionToParse = strings.TrimPrefix(versionToParse, "v")
	}

	parsedRequestVersion, err = ParseVersion(versionToParse)
	if err != nil {
		// 如果解析失败，尝试作为通配符处理（返回最新版本）
		for _, v := range availableVersions {
			normalizedV := strings.ToLower(v)
			if !strings.Contains(normalizedV, "rc") &&
				!strings.Contains(normalizedV, "alpha") &&
				!strings.Contains(normalizedV, "beta") {
				return v, true
			}
		}

		// 如果没有找到稳定版本，返回第一个版本
		return availableVersions[0], true
	}

	var closestVersion string
	var closestVersionParts []int

	for _, v := range availableVersions {
		normalizedV := v
		if stripPrefix != "" && strings.HasPrefix(normalizedV, stripPrefix) {
			normalizedV = strings.TrimPrefix(normalizedV, stripPrefix)
		}

		versionToParse := normalizedV
		if strings.HasPrefix(versionToParse, "v") {
			versionToParse = strings.TrimPrefix(versionToParse, "v")
		}

		vParts, err := ParseVersion(versionToParse)
		if err != nil {
			continue // 跳过无法解析的版本
		}

		// 初始化closestVersion
		if closestVersion == "" {
			closestVersion = v
			closestVersionParts = vParts
			continue
		}

		// 如果当前版本比用户请求的版本小或相等，且比当前找到的最接近版本大，就更新
		result := CompareVersions(vParts, parsedRequestVersion)
		if result <= 0 { // 当前版本 <= 用户请求版本
			result = CompareVersions(vParts, closestVersionParts)
			if result > 0 { // 当前版本 > 已找到的最接近版本
				closestVersion = v
				closestVersionParts = vParts
			}
		}
	}

	if closestVersion != "" {
		return closestVersion, true
	}

	// 4. 如果都找不到，返回第一个非预发布版本
	for _, v := range availableVersions {
		normalizedV := strings.ToLower(v)
		if !strings.Contains(normalizedV, "rc") &&
			!strings.Contains(normalizedV, "alpha") &&
			!strings.Contains(normalizedV, "beta") {
			return v, true
		}
	}

	// 5. 实在找不到，返回第一个版本
	return availableVersions[0], true
}

// CheckURLExists 检查URL是否存在
func CheckURLExists(url string) (bool, error) {
	resp, err := http.Head(url)
	if err != nil {
		return false, err
	}

	// 打印响应状态码
	fmt.Printf("URL响应状态码: %d\n", resp.StatusCode)

	return resp.StatusCode == http.StatusOK, nil
}

// GetNextVersionFromList 从版本列表中获取下一个可用版本
func GetNextVersionFromList(currentVersion string, availableVersions []string) (string, bool) {
	for i, v := range availableVersions {
		if v == currentVersion && i+1 < len(availableVersions) {
			return availableVersions[i+1], true
		}
	}
	return "", false
}

// Min 返回两个整数中较小的一个
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// GetSortedVersions 获取已排序的版本列表（从新到旧）
func GetSortedVersions(versions []string, stripPrefix string) []string {
	result := make([]string, len(versions))
	copy(result, versions)

	SortVersionsDesc(result)
	return result
}

// BuildDownloadURL 构建下载URL
func BuildDownloadURL(baseURL, version, filePattern string) string {
	return fmt.Sprintf(filePattern, baseURL, version)
}
