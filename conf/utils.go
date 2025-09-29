package conf

import "strings"


func removeEmptyStrings(strs []string) []string {
	var result []string
	for _, s := range strs {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, s) 
		}
	}
	return result
}