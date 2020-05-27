package db

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

/**
依据表获得column字段,查询
*/
func GetField(db *Db, table interface{}) []string {
	columns := make([]string, 0)
	re := reflect.ValueOf(table)
	tre := re.Type()
	var t string
BB:
	for {
		switch tre.Kind() {
		case reflect.Slice:
			tre = tre.Elem()
		case reflect.Struct:
			t = tre.Name()
			break BB
		case reflect.String:
			t = table.(string)
			break BB
		case reflect.Ptr:
			tre = tre.Elem()
		default:
			return columns
		}
	}
	sql := fmt.Sprintf("show columns from `%s`", ChangeName(t, 0))
	rows, err := db.Query(sql)
	if err != nil {
		return columns
	}
	var Filed string
	var Type string
	var isNull string
	var Key string
	var Default interface{}
	var AutoIncrement string
	for rows.Next() {
		if err := rows.Scan(&Filed, &Type, &isNull, &Key, &Default, &AutoIncrement); err != nil {
			fmt.Println(err)
			continue
		}
		columns = append(columns, Filed)
	}
	return columns
}

/**
 t==1 是 使用 _  分割字符串
t==0 是 驼峰风格
*/
func ChangeName(str string, t int) string {
	if t == 1 {
		re := regexp.MustCompile("_[a-zA-Z]")
		return re.ReplaceAllStringFunc(str, func(s string) string {
			return strings.ToUpper(strings.Replace(s, "_", "", 1))
		})
	} else {
		re := regexp.MustCompile("[A-Z]")
		str = re.ReplaceAllStringFunc(str, func(s string) string {
			return "_" + strings.ToLower(s)
		})
		return strings.Trim(str, "_")
	}
}

/**
不支持嵌套struct
*/
func AddressStruct(dest interface{}, columns []string) ([]interface{}, error) {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New("亲传递指针")
	}
	addrs := make([]interface{}, 0)
	v = v.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		vf := v.Field(i)
		tf := t.Field(i)
		if tf.Anonymous {
			continue
		}
		if vf.Kind() == reflect.Ptr {
			vf = vf.Elem()
		}
		if vf.Kind() == reflect.Struct && tf.Type.Name() != "Time" {
			continue
		}
		column := strings.Split(tf.Tag.Get("json"), ",")[0]
		if column == "" || column == "-" {
			continue
		}
		for _, col := range columns {
			if col == column {
				addrs = append(addrs, vf.Addr().Interface())
			}
		}
	}
	if len(addrs) != len(columns) {
		return nil, errors.New("scan的长度无法对齐")
	}
	return addrs, nil
}
