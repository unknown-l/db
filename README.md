# db使用说明

## 表连接，分页
```go
db := (&database.Cluster{}).Get(ctx)
// 变量
user := make([]*protobuf.User, 0)
area := make([]*protobuf.User, 0)
var total int32
// 查询
err := db.Ep().Table(&user).Join(&area, "a.area_id=b.id").Filed("a.id,a.name,b.id,b.name").Page(1, 10, &total)
// 打印
fmt.Println(user)
fmt.Println(area)
fmt.Println(err)
```

## 只要主表数据，又使用了连接表
```go
db := (&database.Cluster{}).Get(ctx)
// 变量
user := make([]*protobuf.User, 0)
// 查询
err := db.Ep().Table(&user).JoinName("area", "a.area_id=b.id").Filed("a.id,a.name").Where("a.id<100").Where("b.id=13").Select()
// 打印
fmt.Println(user)
fmt.Println(err)
```

## 单表查询
```go
db := (&database.Cluster{}).Get(ctx)
// 变量
user := protobuf.User{}
// 查询
err := db.Ep().Table(&user).JoinName("area", "a.area_id=b.id").Filed("a.id,a.name").Find()
// 打印
fmt.Println(user)
fmt.Println(err)
```

## 查询数量,group,order
```go
db := (&database.Cluster{}).Get(ctx)
// 变量
user := protobuf.User{}
var total int32
// 查询
err := db.Ep().Table(&user).JoinName("area", "a.area_id=b.id").Group("a.id", "b.id").Order("a.id asc", "b.id desc").Count(&total)
// 打印
fmt.Println(total)
fmt.Println(err)
```

## 查询一列的数组
```go
name := make([]string, 0)
err := (&database.Cluster{}).Ep().Table(&protobuf.User{}).Where("a.id<10").Column("a.name", &name)
// ['', '', '']
```

## 查询单个值
```go
var id int32
err := (&database.Cluster{}).Ep().Table(&protobuf.User{}).Where("a.id<10").Value("a.id", &id)
// 1
```

## in 数组
```go
db.Ep().Table(&user).Where(db.Ep().InInt32("a.id", []int32{})).Where(db.Ep().InString("a.id", []string{}))
```
## where
```go
// "(a in (1, 2)) and (c)"
db.Ep().Table(&user).Where("a in ?", arr).Where("c")

// "(a in (1, 2)) or (c)"
db.Ep().Table(&user).Where("a in ?", arr).WhereOr("c")
```

## 纯原生查询
```go
rows, err := db.Ep().Query(query)
err := db.Ep().QueryRaw(query).Scan()
rows, err := db.Ep().Exec(query)
```

## 执行sql到对象
```go
err := db.Ep().Table(&user).MyQuery(query)
err := db.Ep().Table(&user).MyQueryRaw(query)
```

## 插入单个，返回插入的id，只有赋值的数据会插入
```go
user := protobuf.User{
    Name:                 "123",
    Telephone:            "456",
}
id, _ := (&database.Cluster{}).Ep().Table(&user).Insert()
user.Id = int32(id)
```

## 插入多个，返回影响的条数，只有赋值的数据会插入
```go
user := []*protobuf.User{
		{
			Name:                 "123",
			Telephone:            "456",
		},
		{
			Name:                 "1235",
			Telephone:            "4566",
		},
	}
count, _ := (&database.Cluster{}).Ep().Table(&user).InsertAll()
```

## 更新, 返回影响的条数，只有赋值的数据会更新
```go
user := protobuf.User{
    Name: "123",
    Telephone: "456",
}
count, _ = (&database.Cluster{}).Ep().Table(&user).Where("id=1").Update()
```

## 删除, 返回影响的条数
```go
count, _ = (&database.Cluster{}).Ep().Table(&protobuf.User{}).Where("id=1").Delete()
```