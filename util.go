package db

import (
	"fmt"
)

// InInt32 int32数组转字符串
func InInt32(arg []int32) string {
	if len(arg) == 0 {
		return "\"\""
	}
	result := ""
	for key, item := range arg {
		if key == 0 {
			result += fmt.Sprintf("%d", item)
		} else {
			result += fmt.Sprintf(",%d", item)
		}
	}
	return result
}
func InFloat64(arg []float64) string {
	if len(arg) == 0 {
		return "\"\""
	}
	result := ""
	for key, item := range arg {
		if key == 0 {
			result += fmt.Sprintf("%f", item)
		} else {
			result += fmt.Sprintf(",%f", item)
		}
	}
	return result
}
func InString(arg []string) string {
	if len(arg) == 0 {
		return "\"\""
	}
	result := ""
	for key, item := range arg {
		if key == 0 {
			result += fmt.Sprintf("'%s'", item)
		} else {
			result += fmt.Sprintf(",'%s'", item)
		}
	}
	return result
}
