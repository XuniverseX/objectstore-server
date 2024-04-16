package db

import (
	"fmt"
	mysqlDB "objectstore-server/db/mysql"
)

// OnFileUploadFinished 文件上传成功，保存元数据
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	stmt, err := mysqlDB.DBConn().Prepare(
		"insert ignore into tbl_file(`file_hash`,`file_name`,`filesize`," +
			"`fileaddr`,`status`) values (?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement,", err)
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 {
			// 之前插入过相同的文件
			fmt.Println("This file", filehash, "has been uploaded")
		}
		return true
	}
	return false
}
