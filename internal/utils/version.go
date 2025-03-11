package utils

import (
	"sort"
	"strconv"
	"strings"
)

// 将版本号字符串转换为可比较的整数切片
func parseVersion(version string) []int {
	// 移除可能的'v'前缀
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	parts := strings.Split(version, ".")
	result := make([]int, len(parts))
	for i, part := range parts {
		num, _ := strconv.Atoi(part)
		result[i] = num
	}
	return result
}

// 比较两个版本号
func compareVersions(v1, v2 string) bool {
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// 使用最短的长度进行比较
	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		if parts1[i] != parts2[i] {
			return parts1[i] > parts2[i]
		}
	}

	// 如果前面的部分都相同，较长的版本号较大
	return len(parts1) > len(parts2)
}

// SortVersionsDesc 按版本号降序排序字符串切片
func SortVersionsDesc(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		return compareVersions(versions[i], versions[j])
	})
}

// ParseVersion 将版本号解析为可比较的格式
func ParseVersion(version string) ([]int, error) {
	// 移除可能的'v'前缀
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}

	parts := strings.Split(version, ".")
	result := make([]int, len(parts))
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}
		result[i] = num
	}
	return result, nil
}

// CompareVersions 比较两个版本号，返回:
// -1 如果 v1 < v2
//
//	0 如果 v1 == v2
//	1 如果 v1 > v2
func CompareVersions(v1 []int, v2 []int) int {
	// 使用最短的长度进行比较
	minLen := len(v1)
	if len(v2) < minLen {
		minLen = len(v2)
	}

	for i := 0; i < minLen; i++ {
		if v1[i] < v2[i] {
			return -1
		}
		if v1[i] > v2[i] {
			return 1
		}
	}

	// 如果前面的部分都相同，较长的版本号较大
	if len(v1) < len(v2) {
		return -1
	}
	if len(v1) > len(v2) {
		return 1
	}
	return 0
}

// CompareVersionsStr 比较两个版本号字符串，返回:
// -1 如果 v1 < v2
//
//	0 如果 v1 == v2
//	1 如果 v1 > v2
func CompareVersionsStr(v1, v2 string) int {
	parts1, err1 := ParseVersion(v1)
	parts2, err2 := ParseVersion(v2)

	// 如果解析出错，回退到字符串比较
	if err1 != nil || err2 != nil {
		if v1 == v2 {
			return 0
		}
		if v1 > v2 {
			return 1
		}
		return -1
	}

	return CompareVersions(parts1, parts2)
}
