package util

import "strings"

func ContainsIgnoreCase(list []string, str string) (string, bool) {
	if list == nil {
		return "", false
	}
	str = strings.ToLower(str)
	for _, s := range list {
		if str == strings.ToLower(s) {
			return s, true
		}
	}
	return "", false
}

func Contains(list []string, str string) (string, bool) {
	if list == nil {
		return "", false
	}
	for _, s := range list {
		if str == s {
			return s, true
		}
	}
	return "", false
}
