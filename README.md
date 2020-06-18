# 支持mysql的数据库工具
- 安装 ( install )
```go
import "github.com/unknown-l/db"
```
- 连接数据库 ( connect mysql )
```go
database := map[string][]string{
    "db1": {"username:password@tcp(host:port)/database", "username:password@tcp(host:port)/database"},
    "db2": {"username:password@tcp(host:port)/database", "username:password@tcp(host:port)/database"},
}
cluster, _ := db.NewCluster(database)
db1 := cluster.Db("db1")
```
- 查询 ( select )
```go
type User struct {
    Id int32
    Name string
    RoleId int32
}
user := User{}
ids := []int32{1, 2}
_ = cluster.Db("db1").Table(&user).Filed("id,name").Where("id in ? and name = ?", ids, "a").Find()
```
- 分页 ( page )
```go
user := make([]*User, 0)
var page int32 = 1
var limit int32 = 10
var total int32 = 0
err := cluster.Db("db1").Table(&user).Filed("name").Page(page, limit, &total)
```
- 左连接查询 ( left join )
```go
type Role struct {
    Id int32
    Name string
}
user := make([]*User, 0)
role := make([]*Role, 0)
err := cluster.Db("db1").Table(&user).Filed("a.name,b.name").
    Join(&role, "a.role_id=b.id").
    JoinName("role", "a.role_id=c.id").
    Select()
```
- 查询单个值 ( select one row one column value )
```go
name := ""
err := cluster.Db("db1").Table(&User{}).Where("id=1").Value("name", &name)
```
- 查询单个列的多个值 ( select multiple rows one column value )
```go
name := make([]string, 0)
err := cluster.Db("db1").Table(&User{}).Where("id=1").Column("name", &name)
```
- 新增单条记录 ( insert one record )
```go
user := User{Name: "username"}
recordId, err := cluster.Db("db1").Table(&user).Insert()
```
- 新增多条记录 ( insert multiple records )
```go
users := []*User{{Name: "username1"}, {Name: "username2"}}
recordCount, err := cluster.Db("db1").Table(&users).InsertAll()
```
- 更新 - 修改对象中非空`!= 0, ""`的值 ( update - modify only values that are not equal to 0, "" )
```go
user := User{Id: 1, Name: "username"}
recordCount, err := cluster.Db("db1").Table(&user).Where("id=?", user.Id).Update()
```
- 更新 - 指定Field，0值也会更新 ( update - set fields to update )
```go
user := User{Id: 1, Name: ""}
recordCount, err := cluster.Db("db1").Table(&user).Filed("name").Where("id=?", user.Id).Update()
```