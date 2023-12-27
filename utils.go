package fesgo

import "strings"

func SubStringLast(str string, substring string) string {
	index := strings.Index(str, substring)
	if index < 0 {
		return ""
	}

	return str[index+len(substring):]
}
