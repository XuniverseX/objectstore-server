package db

import (
	"fmt"
	mysqlDB "objectstore-server/db/mysql"
)

// UserSignUp 通过用户名密码注册用户
func UserSignUp(username string, password string) bool {
	stmt, err := mysqlDB.DBConn().Prepare(
		"insert into tbl_user(`user_name`, `user_pwd`) values(?,?)")
	if err != nil {
		fmt.Println("Failed to insert,", err)
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Println("Failed to insert,", err)
	}

	if rowsAffected, err := ret.RowsAffected(); err == nil && rowsAffected > 0 {
		return true
	}
	return false
}
