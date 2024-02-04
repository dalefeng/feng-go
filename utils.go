package fesgo

import (
	"strings"
	"unicode"
)

func SubStringLast(str string, substring string) string {
	index := strings.Index(str, substring)
	if index < 0 {
		return ""
	}

	return str[index+len(substring):]
}

func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}
