package db

import (
	"fmt"
	"log"
	mdb "objectstore-server/db/mysql"
)

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// UserSignUp 通过用户名密码注册用户
func UserSignUp(username string, password string) bool {
	stmt, err := mdb.DBConn().Prepare(
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

func UserLogin(username string, enc_pwd string) bool {
	stmt, err := mdb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err)
		return false
	}

	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err)
		return false
	} else if rows == nil {
		fmt.Println("username not found:")
		return false
	}

	pRows := mdb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == enc_pwd {
		return true
	}
	return false
}

// UpdateToken 更新用户登录token
func UpdateToken(username string, token string) bool {
	stmt, err := mdb.DBConn().Prepare(
		"replace into tbl_user_token(`user_name`, `user_token`) values(?,?)")
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// GetUserInfo get user	by username
func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := mdb.DBConn().Prepare(
		"select user_name, signup_at from tbl_user where user_name=? limit 1",
	)
	if err != nil {
		log.Println(err)
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		log.Println(err)
		return user, err
	}
	return user, nil
}
