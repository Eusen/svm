package utils

import (
	"bytes"
	"os/exec"
)

// RunCommand 执行命令并返回输出
func RunCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
} 