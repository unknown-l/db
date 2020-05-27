package db

import "reflect"

type column struct {
	className string        // 类中的名称
	kind      string        // 类型
	value     reflect.Value // 值
}
