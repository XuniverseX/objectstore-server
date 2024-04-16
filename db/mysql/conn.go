package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:root@tcp(127.0.0.1:3307)/fileserver?charset=utf8mb4")
	db.SetMaxIdleConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect mysql,", err)
		os.Exit(1)
	}
}

// DBConn 返回数据库连接对象
func DBConn() *sql.DB {
	return db
}
