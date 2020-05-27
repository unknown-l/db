package db

import "reflect"

type table struct {
	name      string               // 数据库名
	alias     string               // alias
	join      string               // 连接
	on        string               // 连接条件
	obj       interface{}          // 原始值
	column    map[string]*column   // 字段类型
	objVal    []map[string]*column // 字段的值
	arrValue  reflect.Value        // Value
	itemValue reflect.Value        // 数组中结构体
	itemType  reflect.Type         // Type
}

const tableMasterAlias = "a"

var tableSlaveAlias = []string{"b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q"}
