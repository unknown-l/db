package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"strings"
)

const commandSelect = "select"
const commandInsert = "insert"
const commandUpdate = "update"
const commandDelete = "delete"

type Db struct {
	master  *table
	slave   map[string]*table
	field   *field
	where   []*where
	order   string
	group   string
	have    string
	limit   string
	command string
	db      []*sql.DB
	dbTx    *sql.Tx
}

// ==============================初始化表格===========================
/* table */
func (d *Db) Table(m interface{}) *Db {
	newDb := d.newQuery()
	newDb.master.obj = m
	newDb.master.alias = tableMasterAlias
	newDb.command = commandSelect
	newDb.reflectColumn(newDb.master, m)
	return newDb
}
func (d *Db) Join(m interface{}, on string, args ...interface{}) *Db {
	slave := &table{column: map[string]*column{}}
	slave.alias = fmt.Sprintf("%s", tableSlaveAlias[len(d.slave)])
	slave.join = "left join"
	slave.obj = m
	slave.on = d.FormatMarkStr(on, args...)
	d.reflectColumn(slave, m)
	d.slave[slave.alias] = slave
	return d
}
func (d *Db) InnerJoin(m interface{}, on string, args ...interface{}) *Db {
	slave := &table{column: map[string]*column{}}
	slave.alias = fmt.Sprintf("%s", tableSlaveAlias[len(d.slave)])
	slave.join = "inner join"
	slave.obj = m
	slave.on = d.FormatMarkStr(on, args...)
	d.reflectColumn(slave, m)
	d.slave[slave.alias] = slave
	return d
}

/* 例如需要从user查询数据到统计对象TableName(&count{}, "user") */
func (d *Db) TableName(m interface{}, tableName string) *Db {
	newDb := d.newQuery()
	newDb.master.obj = m
	newDb.master.alias = tableMasterAlias
	newDb.command = commandSelect
	newDb.reflectColumn(newDb.master, m)
	newDb.master.name = tableName
	return newDb
}

/* 不需要解析数据到关联表TableName("area", on) */
func (d *Db) JoinName(tableName string, on string, args ...interface{}) *Db {
	slave := &table{column: map[string]*column{}}
	slave.alias = fmt.Sprintf("%s", tableSlaveAlias[len(d.slave)])
	slave.join = "left join"
	slave.name = tableName
	slave.on = d.FormatMarkStr(on, args...)
	d.slave[slave.alias] = slave
	return d
}

// ==============================where===========================
func (d *Db) Where(query string, args ...interface{}) *Db {
	return d.wherefFormat("and", "and", query, args...)
}
func (d *Db) WhereOr(query string, args ...interface{}) *Db {
	return d.wherefFormat("or", "and", query, args...)
}
func (d *Db) wherefFormat(left string, inner string, query string, args ...interface{}) *Db {
	buf := d.FormatMarkStr(query, args...)
	//
	whereItem := &where{join: left, combine: inner, item: make([]string, 0)}
	whereItem.item = append(whereItem.item, buf)
	d.where = append(d.where, whereItem)
	return d
}

// ==============================查询===========================
func (d *Db) Select() error {
	query := d.Sql()
	rows, err := d.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	d.reflectArrType()
	values := d.scanArr()
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return err
		}
		d.setScanArrValue(values)
	}
	return nil
}
func (d *Db) Find() error {
	d.limit = "1"
	query := d.Sql()
	row := d.QueryRow(query)
	values := d.scanArr()
	if err := row.Scan(values...); err != nil {
		return err
	}
	d.setScanItemValue(values)
	return nil
}

/**
如果查询单个没有
是一个正常的错误时使用
不在单个处理
*/
func (d *Db) Get() error {
	d.limit = "1"
	query := d.Sql()
	row := d.QueryRow(query)
	values := d.scanArr()
	if err := row.Scan(values...); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	d.setScanItemValue(values)
	return nil
}
func (d *Db) Count(count *int32) error {
	//  限制
	d.order = ""
	d.limit = ""
	d.field.string = "null"
	// 查询
	query := fmt.Sprintf("select count(*) from (%s) a", d.Sql())
	nullCount := sql.NullInt32{}
	if err := d.QueryRow(query).Scan(&nullCount); err != nil {
		return err
	}
	if nullCount.Valid {
		*count = nullCount.Int32
	}
	return nil
}
func (d *Db) Sum(field string, sum interface{}) error {
	//  限制
	d.field.Set(fmt.Sprintf("sum(%s)", field))
	d.order = ""
	d.limit = ""
	// 查询
	query := d.Sql()
	// 反射
	sumVal := reflect.ValueOf(sum)
	sumType := sumVal.Elem().Type().Name()
	var nullSum interface{}
	if sumType == "int32" {
		nullSum = &sql.NullInt32{}
	} else if sumType == "float64" {
		nullSum = &sql.NullFloat64{}
	}
	if err := d.QueryRow(query).Scan(nullSum); err != nil {
		return err
	}
	// 取值
	if sumType == "int32" && nullSum.(*sql.NullInt32).Valid {
		sumVal.Elem().Set(reflect.ValueOf(nullSum.(*sql.NullInt32).Int32))
	} else if sumType == "float64" && nullSum.(*sql.NullFloat64).Valid {
		sumVal.Elem().Set(reflect.ValueOf(nullSum.(*sql.NullFloat64).Float64))
	}
	return nil
}
func (d *Db) Page(page int32, limit int32, total *int32) error {

	// 查询列表
	d.limit = fmt.Sprintf("%d,%d", (page-1)*limit, limit)
	if err := d.Select(); err != nil {
		return err
	}
	// 查询数量
	if err := d.Count(total); err != nil {
		return err
	}
	return nil
}
func (d *Db) Column(field string, result interface{}) error {
	d.field.Set(field)
	query := d.Sql()
	rows, err := d.Query(query)
	if err != nil {
		return err
	}
	// 反射构建数组
	arrValue := reflect.ValueOf(result).Elem()
	arrValue.Set(reflect.MakeSlice(arrValue.Type(), 0, 0))
	itemType := arrValue.Type().Elem()
	var resultItem interface{}

	for rows.Next() {
		if itemType.Name() == "int32" {
			resultItem = &sql.NullInt32{}
		} else if itemType.Name() == "string" {
			resultItem = &sql.NullString{}
		} else if itemType.Name() == "float" {
			resultItem = &sql.NullFloat64{}
		} else if itemType.Name() == "int64" {
			resultItem = &sql.NullInt64{}
		}
		if err = rows.Scan(resultItem); err != nil {
			return err
		}
		itemValue := reflect.New(itemType).Elem()
		if itemType.Name() == "int32" && resultItem.(*sql.NullInt32).Valid {
			itemValue.Set(reflect.ValueOf(resultItem.(*sql.NullInt32).Int32))
		} else if itemType.Name() == "string" && resultItem.(*sql.NullString).Valid {
			itemValue.Set(reflect.ValueOf(resultItem.(*sql.NullString).String))
		} else if itemType.Name() == "float64" && resultItem.(*sql.NullFloat64).Valid {
			itemValue.Set(reflect.ValueOf(resultItem.(*sql.NullFloat64).Float64))
		} else if itemType.Name() == "int64" && resultItem.(*sql.NullInt64).Valid {
			itemValue.Set(reflect.ValueOf(resultItem.(*sql.NullInt64).Int64))
		}
		arrValue.Set(reflect.Append(arrValue, itemValue))
	}
	return nil
}
func (d *Db) Value(field string, result interface{}) error {
	d.field.Set(field)
	d.limit = "1"
	query := d.Sql()
	if err := d.QueryRow(query).Scan(result); err != nil {
		return err
	}
	return nil
}
func (d *Db) MyQuery(query string) error {
	rows, err := d.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	d.reflectArrType()
	values := d.scanArr()
	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return err
		}
		d.setScanArrValue(values)
	}
	return nil
}
func (d *Db) MyQueryRaw(query string) error {
	d.limit = "1"
	row := d.QueryRow(query)
	values := d.scanArr()
	if err := row.Scan(values...); err != nil {
		return err
	}
	d.setScanItemValue(values)
	return nil
}

// ==============================插入或者更新===========================
func (d *Db) Insert() (int64, error) {
	d.command = commandInsert
	query := d.Sql()
	result, err := d.Exec(query)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return id, err
}
func (d *Db) InsertAll() (int64, error) {
	d.command = commandInsert
	query := d.Sql()
	result, err := d.Exec(query)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	return count, err
}
func (d *Db) Update() (int64, error) {
	d.command = commandUpdate
	query := d.Sql()
	result, err := d.Exec(query)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	return count, err
}
func (d *Db) Delete() (int64, error) {
	d.command = commandDelete
	query := d.Sql()
	result, err := d.Exec(query)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	return count, err
}

// ==============================生成sql===========================
func (d *Db) Sql() string {
	switch d.command {
	case commandSelect:
		return d.selectSql()
	case commandInsert:
		return d.insertSql()
	case commandUpdate:
		return d.updateSql()
	case commandDelete:
		return d.deleteSql()
	}
	return ""
}
func (d *Db) selectSql() string {
	selectTmp := "select %s from %s %s"
	joinTmp := "%s %s %s on %s"
	groupTmp := "group by %s"
	haveTmp := "having %s"
	orderTmp := "order by %s"
	limitTmp := "limit %s"
	query := make([]string, 0)
	// master and field
	query = append(query, fmt.Sprintf(selectTmp, d.field.string, d.master.name, d.master.alias))
	// join
	joinArr := make([]string, 0)
	for _, slave := range d.slave {
		joinArr = append(joinArr, fmt.Sprintf(joinTmp, slave.join, slave.name, slave.alias, slave.on))
	}
	// slave
	if len(joinArr) != 0 {
		query = append(query, strings.Join(joinArr, " "))
	}
	// where
	whereStr := d.whereSql()
	if whereStr != "" {
		query = append(query, whereStr)
	}
	// group
	if d.group != "" {
		query = append(query, fmt.Sprintf(groupTmp, d.group))
	}
	// have
	if d.have != "" {
		query = append(query, fmt.Sprintf(haveTmp, d.have))
	}
	// order
	if d.order != "" {
		query = append(query, fmt.Sprintf(orderTmp, d.order))
	}
	// limit
	if d.limit != "" {
		query = append(query, fmt.Sprintf(limitTmp, d.limit))
	}
	return strings.Join(query, " ")
}
func (d *Db) whereSql() string {
	whereTmp := "where %s"
	query := ""
	// where
	whereStrArr := make([]string, 0)
	for _, whereItem := range d.where {
		if len(whereStrArr) == 0 {
			whereStrArr = append(whereStrArr, fmt.Sprintf("(%s)", strings.Join(whereItem.item, fmt.Sprintf(" %s ", whereItem.combine))))
		} else {
			whereStrArr = append(whereStrArr, fmt.Sprintf("%s (%s)", whereItem.join, strings.Join(whereItem.item, fmt.Sprintf(" %s ", whereItem.combine))))
		}
	}
	if len(whereStrArr) != 0 {
		query = fmt.Sprintf(whereTmp, strings.Join(whereStrArr, " "))
	}
	return query
}
func (d *Db) insertSql() string {
	// table_name, column, value
	insertTmp := "INSERT INTO %s (%s) VALUES %s"
	query := make([]string, 0)
	// 这里要循环两次，找出所有不是空的column
	column := make([]string, 0)
	for masterColumnName, masterColumnItem := range d.master.column {
		if !masterColumnItem.value.IsZero() {
			column = append(column, masterColumnName)
		}
	}
	// 获取有值的value
	valRow := make([]string, 0)
	for _, objVal := range d.master.objVal {
		valCol := make([]string, 0)
		for _, columnName := range column {
			// 都没有值就跳过
			if d.master.column[columnName].value.IsZero() {
				continue
			}
			// 插入值
			if objVal[columnName].kind == "int32" {
				valCol = append(valCol, fmt.Sprintf("%d", objVal[columnName].value.Interface()))
			} else if objVal[columnName].kind == "string" {
				valCol = append(valCol, fmt.Sprintf("'%s'", objVal[columnName].value.Interface()))
			} else if objVal[columnName].kind == "float64" {
				valCol = append(valCol, fmt.Sprintf("%f", objVal[columnName].value.Interface()))
			} else if objVal[columnName].kind == "int64" {
				valCol = append(valCol, fmt.Sprintf("%d", objVal[columnName].value.Interface()))
			}
		}
		valRow = append(valRow, fmt.Sprintf("(%s)", strings.Join(valCol, ",")))
	}
	// tmp
	query = append(query, fmt.Sprintf(insertTmp, d.master.name, strings.Join(column, ","), strings.Join(valRow, ",")))
	return strings.Join(query, " ")
}
func (d *Db) updateSql() string {
	// table_name, column, value
	updateTmp := "UPDATE %s %s SET %s"
	query := make([]string, 0)
	// 获取有值的column和value
	column := make([]string, 0)
	for _, objVal := range d.master.objVal {
		for objItemKey, objItemVal := range objVal {
			if objItemVal.kind == "int32" && !objItemVal.value.IsZero() {
				column = append(column, fmt.Sprintf("%s.%s=%d", d.master.alias, objItemKey, objItemVal.value.Interface()))
			} else if objItemVal.kind == "string" && !objItemVal.value.IsZero() {
				column = append(column, fmt.Sprintf("%s.%s='%s'", d.master.alias, objItemKey, objItemVal.value.Interface()))
			} else if objItemVal.kind == "float64" && !objItemVal.value.IsZero() {
				column = append(column, fmt.Sprintf("%s.%s=%f", d.master.alias, objItemKey, objItemVal.value.Interface()))
			} else if objItemVal.kind == "int64" && !objItemVal.value.IsZero() {
				column = append(column, fmt.Sprintf("%s.%s=%d", d.master.alias, objItemKey, objItemVal.value.Interface()))
			}
		}
	}
	// deleteTmp
	query = append(query, fmt.Sprintf(updateTmp, d.master.name, d.master.alias, strings.Join(column, ",")))
	// where
	whereStr := d.whereSql()
	if whereStr != "" {
		query = append(query, whereStr)
	}
	return strings.Join(query, " ")
}
func (d *Db) deleteSql() string {
	// table_name, column, value
	deleteTmp := "DELETE FROM %s"
	query := make([]string, 0)
	// deleteTmp
	query = append(query, fmt.Sprintf(deleteTmp, d.master.name))
	// where
	whereStr := d.whereSql()
	if whereStr != "" {
		query = append(query, whereStr)
	}
	return strings.Join(query, " ")
}

// ==============================其他===========================
func (d *Db) Group(group ...string) *Db {
	d.group = strings.Join(group, ",")
	return d
}
func (d *Db) Have(have string) *Db {
	d.have = have
	return d
}
func (d *Db) Order(order ...string) *Db {
	d.order = strings.Join(order, ",")
	return d
}
func (d *Db) Filed(filed string) *Db {
	d.field.Set(filed)
	return d
}
func (d *Db) FormatMarkStr(query string, args ...interface{}) string {
	buf := make([]byte, 0)
	argPos := 0
	for i := 0; i < len(query); i++ {
		// 没找到? 直接返回
		q := strings.IndexByte(query[i:], '?')
		if q == -1 {
			buf = append(buf, query[i:]...)
			break
		}
		// 有？替换args
		buf = append(buf, query[i:i+q]...)
		i += q
		arg := args[argPos]
		argPos++
		// 反射arg
		argsTypeOf := reflect.TypeOf(arg)
		if argsTypeOf.Kind() == reflect.Int32 || argsTypeOf.Kind() == reflect.Int64 || argsTypeOf.Kind() == reflect.Int {
			buf = append(buf, fmt.Sprintf("%d", arg)[0:]...)
		} else if argsTypeOf.Kind() == reflect.Float64 {
			buf = append(buf, fmt.Sprintf("%f", arg)[0:]...)
		} else if argsTypeOf.Kind() == reflect.String {
			buf = append(buf, fmt.Sprintf("'%s'", arg)[0:]...)
		} else if argsTypeOf.Kind() == reflect.Slice {
			argsValueType := argsTypeOf.Elem().Kind()
			buf = append(buf, '(')
			if argsValueType == reflect.Int32 || argsValueType == reflect.Int64 || argsValueType == reflect.Int {
				buf = append(buf, InInt32(arg.([]int32))[0:]...)
			} else if argsValueType == reflect.Float64 {
				buf = append(buf, InFloat64(arg.([]float64))[0:]...)
			} else if argsValueType == reflect.String {
				buf = append(buf, InString(arg.([]string))[0:]...)
			}
			buf = append(buf, ')')
		}
	}
	return string(buf)
}

// ==============================一些辅助函数===========================
/* 生成一个常数的查询列表 */
func (d *Db) StringArrTable(data []string, name string) string {
	tableArr := make([]string, 0)
	for _, item := range data {
		tableArr = append(tableArr, fmt.Sprintf("SELECT '%s' as %s", item, name))
	}
	return fmt.Sprintf("(%s)", strings.Join(tableArr, " UNION "))
}

/* 帕斯卡命名(go 变量)转下划线(数据库字段) */
func (d *Db) CamelCase(s string) string {
	// 空字符串
	if s == "" {
		return ""
	}
	// 单个字符串
	length := len(s)
	if length == 1 {
		return strings.ToLower(s)
	}
	// 多个字符串
	array := make([]string, 0)
	arrayItem := make([]byte, 0)
	for index := 0; index < length; index++ {
		if 'A' <= s[index] && s[index] <= 'Z' && index != 0 {
			array = append(array, string(arrayItem))
			arrayItem = make([]byte, 0)
		}
		arrayItem = append(arrayItem, s[index])
	}
	array = append(array, string(arrayItem))
	return strings.ToLower(strings.Join(array, "_"))
}

// ==============================Find用到的反射===========================
/*  扫描到的值赋值到对象中 */
func (d *Db) setScanItemValue(value []interface{}) {
	// 取值
	for fieldKey, fieldItem := range d.field.item {
		table := &table{}
		if fieldItem.alias == tableMasterAlias {
			table = d.master
		} else {
			table = d.slave[fieldItem.alias]
		}
		if table.column[fieldItem.name].kind == "int32" {
			if value[fieldKey].(*sql.NullInt32).Valid {
				reflect.ValueOf(table.obj).Elem().FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullInt32).Int32))
			}
		} else if table.column[fieldItem.name].kind == "string" {
			if value[fieldKey].(*sql.NullString).Valid {
				reflect.ValueOf(table.obj).Elem().FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullString).String))
			}
		} else if table.column[fieldItem.name].kind == "float64" {
			if value[fieldKey].(*sql.NullFloat64).Valid {
				reflect.ValueOf(table.obj).Elem().FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullFloat64).Float64))
			}
		} else if table.column[fieldItem.name].kind == "int64" {
			if value[fieldKey].(*sql.NullInt64).Valid {
				reflect.ValueOf(table.obj).Elem().FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullInt64).Int64))
			}
		}
	}
}

// ==============================Select用到的反射===========================
/* 生成数组和类型 */
func (d *Db) reflectArrType() {
	d.master.arrValue = reflect.ValueOf(d.master.obj).Elem()
	d.master.arrValue.Set(reflect.MakeSlice(d.master.arrValue.Type(), 0, 0))
	d.master.itemType = d.master.arrValue.Type().Elem().Elem()
	for key, slave := range d.slave {
		if slave.obj != nil {
			slave.arrValue = reflect.ValueOf(d.slave[key].obj).Elem()
			d.slave[key].arrValue.Set(reflect.MakeSlice(d.slave[key].arrValue.Type(), 0, 0))
			slave.itemType = d.slave[key].arrValue.Type().Elem().Elem()
		}
	}
}

/* 扫描到的值赋值到数组中 */
func (d *Db) setScanArrValue(value []interface{}) {
	// 创建新对象
	d.master.itemValue = reflect.New(d.master.itemType).Elem()
	for key, slave := range d.slave {
		if slave.obj != nil {
			slave.itemValue = reflect.New(d.slave[key].itemType).Elem()
		}
	}
	// 取值
	for fieldKey, fieldItem := range d.field.item {
		table := &table{}
		if fieldItem.alias == tableMasterAlias {
			table = d.master
		} else {
			table = d.slave[fieldItem.alias]
		}
		if table.column[fieldItem.name].kind == "int32" {
			if value[fieldKey].(*sql.NullInt32).Valid {
				table.itemValue.FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullInt32).Int32))
			}
		} else if table.column[fieldItem.name].kind == "string" {
			if value[fieldKey].(*sql.NullString).Valid {
				table.itemValue.FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullString).String))
			}
		} else if table.column[fieldItem.name].kind == "float64" {
			if value[fieldKey].(*sql.NullFloat64).Valid {
				table.itemValue.FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullFloat64).Float64))
			}
		} else if table.column[fieldItem.name].kind == "int64" {
			if value[fieldKey].(*sql.NullInt64).Valid {
				table.itemValue.FieldByName(table.column[fieldItem.name].className).Set(reflect.ValueOf(value[fieldKey].(*sql.NullInt64).Int64))
			}
		}
	}
	// 赋值
	d.master.arrValue.Set(reflect.Append(d.master.arrValue, d.master.itemValue.Addr()))
	for key, slave := range d.slave {
		if slave.obj != nil {
			d.slave[key].arrValue.Set(reflect.Append(slave.arrValue, d.slave[key].itemValue.Addr()))
		}
	}
}

/* 生成scan用的数组 */
func (d *Db) scanArr() []interface{} {
	scanVal := make([]interface{}, len(d.field.item))
	for fieldKey, field := range d.field.item {
		fieldKind := ""
		if field.alias == tableMasterAlias {
			fieldKind = d.master.column[field.name].kind
		} else {
			fieldKind = d.slave[field.alias].column[field.name].kind
		}
		if fieldKind == "int32" {
			scanVal[fieldKey] = &sql.NullInt32{}
		} else if fieldKind == "string" {
			scanVal[fieldKey] = &sql.NullString{}
		} else if fieldKind == "float64" {
			scanVal[fieldKey] = &sql.NullFloat64{}
		} else if fieldKind == "int64" {
			scanVal[fieldKey] = &sql.NullInt64{}
		}
	}
	return scanVal
}

// ==============================Table初始化反射获取列=======================
// 反射获取column
func (d *Db) reflectColumn(t *table, m interface{}) {
	refVal := reflect.ValueOf(m).Elem()
	refType := refVal.Type()
	if refType.Kind() == reflect.Slice {
		refType = refType.Elem().Elem()
	}
	t.name = d.CamelCase(refType.Name())
	// 取列,记录对应列
	names := make([]string, 0)
	className := make([]string, 0)
	kind := make([]string, 0)
	for i := 0; i < refType.NumField(); i++ {
		name := d.CamelCase(refType.Field(i).Name)
		t.column[name] = &column{className: refType.Field(i).Name, kind: refType.Field(i).Type.Name(), value: reflect.ValueOf(0)}
		//
		names = append(names, name)
		className = append(className, t.column[name].className)
		kind = append(kind, t.column[name].kind)
	}
	// 取数组值
	if refVal.Type().Kind() == reflect.Slice {
		for i := 0; i < refVal.Len(); i++ {
			refValItem := refVal.Index(i).Elem()
			objVal := map[string]*column{}
			for j := 0; j < refValItem.NumField(); j++ {
				objVal[names[j]] = &column{className: className[j], kind: kind[j], value: refValItem.Field(j)}
				if !objVal[names[j]].value.IsZero() {
					t.column[names[j]].value = objVal[names[j]].value
				}
			}
			t.objVal = append(t.objVal, objVal)
		}
	} else if refVal.Type().Kind() == reflect.Struct {
		objVal := map[string]*column{}
		refValItem := refVal
		for j := 0; j < refValItem.NumField(); j++ {
			objVal[names[j]] = &column{className: className[j], kind: kind[j], value: refValItem.Field(j)}
			if !objVal[names[j]].value.IsZero() {
				t.column[names[j]].value = objVal[names[j]].value
			}
		}
		t.objVal = append(t.objVal, objVal)
	}
}

// ==============================原生数据库查询==============================
/* 获取数据库 */
func (d *Db) Init() *Db {
	d = &Db{
		master: &table{
			column: map[string]*column{},
			objVal: make([]map[string]*column, 0),
		},
		slave: map[string]*table{},
		field: &field{
			item: make([]*fieldItem, 0),
		},
		where: make([]*where, 0),
		db:    make([]*sql.DB, 0),
		dbTx:  nil,
	}
	return d
}
func (d *Db) Master() *sql.DB {
	return d.db[0]
}
func (d *Db) Slave() *sql.DB {
	return d.db[len(d.db)-1]
}

// 每次调用table，初始化条件
func (d *Db) newQuery() *Db {
	db := *d
	db.master = &table{
		column: map[string]*column{},
		objVal: make([]map[string]*column, 0),
	}
	db.slave = map[string]*table{}
	db.field = &field{
		item: make([]*fieldItem, 0),
	}
	db.where = make([]*where, 0)
	db.order = ""
	db.group = ""
	db.have = ""
	db.limit = ""
	db.command = ""
	return &db
}

/* 事务 */
func (d *Db) IsTrans() bool {
	return d.dbTx != nil
}
func (d *Db) StartTrans() error {
	if d.IsTrans() {
		return nil
	}
	dbTx, err := d.Master().Begin()
	if err != nil {
		return err
	}
	d.dbTx = dbTx
	return nil
}
func (d *Db) Commit() error {
	if d.IsTrans() {
		return d.dbTx.Commit()
	}
	return nil
}
func (d *Db) Rollback() error {
	if d.IsTrans() {
		return d.dbTx.Rollback()
	}
	return nil
}

/* 查询 */
func (d *Db) QueryRow(query string, args ...interface{}) *sql.Row {

	return d.Slave().QueryRow(query, args...)
}
func (d *Db) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.Slave().Query(query, args...)
}

/* 修改 */
func (d *Db) Exec(query string, args ...interface{}) (sql.Result, error) {
	if d.IsTrans() {
		return d.dbTx.Exec(query, args...)
	}
	return d.Master().Exec(query, args...)
}

/* 关闭 */
func (d *Db) Close() error {
	for _, db := range d.db {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}
