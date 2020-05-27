package db

import (
	"context"
	"database/sql"
)

// Ctx 上下文传递的key
const Ctx = "DatabaseCtx"

// cluster 数据库集群类
type cluster struct {
	db map[string]*Db
}

// 新建集群
func NewCluster(dbsMap map[string][]string) (*cluster, error) {
	c := &cluster{db: make(map[string]*Db, 0)}
	for dbKey, dbs := range dbsMap {
		instance := Db{db: make([]*sql.DB, 0)}
		for _, dbStr := range dbs {
			db, err := sql.Open("mysql", dbStr)
			if err != nil {
				return c, err
			}
			instance.db = append(instance.db, db)
		}
		c.db[dbKey] = &instance
	}
	return c, nil
}

// Db 获取连接
func (c *cluster) Db(key string) *Db {
	return c.db[key]
}

/* Done 根据错误判断事务处理 */
func (c *cluster) Done(errParam error) error {
	// 判断mysql的回滚
	if errParam == nil {
		// 成功提交
		for _, db := range c.db {
			if err := db.Commit(); err != nil {
				return err
			}
		}
	} else {
		// 失败回滚
		for _, db := range c.db {
			if err := db.Rollback(); err != nil {
				return err
			}
		}
	}
	return nil
}

/* 关闭连接 */
func (c *cluster) Close() error {
	for _, db := range c.db {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Context 获取上下文传递的集群
func (c *cluster) Context(ctx context.Context) *cluster {
	return ctx.Value(Ctx).(*cluster)
}
