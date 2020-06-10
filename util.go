package db

import (
	"fmt"
)

// InInt int数组转字符串
func InInt(arg []int) string {
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

// InInt64 int64数组转字符串
func InInt64(arg []int64) string {
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
