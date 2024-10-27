package helpers

import "strings"

func InArray(tofind string, array []string) bool {
	for _, element := range array {
		if strings.EqualFold(element, tofind) {
			return true
		}
	}
	return false
}
