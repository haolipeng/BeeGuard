package main

import (
	"bytes"
	"strings"
)

// cstring 将C字符串（以\0结尾）转换为Go字符串
func cstring(b []byte) string {
	n := bytes.IndexByte(b, 0)
	if n == -1 {
		n = len(b)
	}
	return string(b[:n])
}

// argsString 处理命令行参数：将NULL字节分隔的多个参数转换为空格分隔的字符串
func argsString(b []byte) string {
	end := len(b)
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			allZero := true
			for j := i; j < len(b) && j < i+4; j++ {
				if b[j] != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				end = i
				break
			}
		}
	}
	result := make([]byte, end)
	copy(result, b[:end])
	for i := 0; i < len(result); i++ {
		if result[i] == 0 {
			result[i] = ' '
		}
	}
	return string(bytes.TrimRight(result, " "))
}

// extractLeadingCommand 从匹配模式中提取前导命令名
// 例如: "rm\s+.*-rf" → "rm", "cat\s+.*/etc/shadow" → "cat", "chmod\s+.*777" → "chmod"
// 遇到正则元字符或空白转义时停止
func extractLeadingCommand(pattern string) string {
	var cmd strings.Builder
	for _, ch := range pattern {
		if ch == '\\' || ch == '(' || ch == '[' || ch == '.' || ch == '*' ||
			ch == '+' || ch == '?' || ch == '{' || ch == '|' || ch == '^' || ch == '$' {
			break
		}
		if ch == ' ' || ch == '\t' {
			break
		}
		cmd.WriteRune(ch)
	}
	return cmd.String()
}
