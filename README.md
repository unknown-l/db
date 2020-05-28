# 支持mysql的数据库工具
- 安装
```go
import "github.com/unknown-l/db"
```
- 连接数据库
```go
database := map[string][]string{
    "db1": {"username:password@tcp(host:port)/database", "username:password@tcp(host:port)/database"},
    "db2": {"username:password@tcp(host:port)/database", "username:password@tcp(host:port)/database"},
}
cluster, _ := db.NewCluster(database)
db1 := cluster.Db("db1")
```
- 查询
```go
type User struct {
    Id int32
    Name string
}
user := User{}
ids := []int32{1, 2}
_ = cluster.Db("db1").Table(&user).Filed("id,name").Where("id in ? and name = ?", ids, "a").Find()
```