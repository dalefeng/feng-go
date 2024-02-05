package fesgo

import (
	"net/url"
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

// ParseParamsMap 解析查询字符串转换为Map
func ParseParamsMap(values url.Values) map[string]map[string]string {
	m := make(map[string]map[string]string)
	for key, value := range values {
		keys := strings.Split(key, "[")
		if len(keys) != 2 {
			continue
		}
		mainKey := keys[0]
		subKey := strings.TrimRight(keys[1], "]")
		if m[mainKey] == nil {
			m[mainKey] = make(map[string]string)
		}
		m[mainKey][subKey] = value[len(value)-1]
	}

	return m
}
