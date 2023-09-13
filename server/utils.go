package main

import "strings"

// ExtractString 접두사와 공백제거 후 추출
func ExtractString(s, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(s, prefix))
}
