package db

import (
	"fmt"
	"log"
	mdb "objectstore-server/db/mysql"
	"time"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished 上传完成时插入用户文件表
func OnUserFileUploadFinished(username, fileHash, fileName string, fileSize int64) bool {
	stmt, err := mdb.DBConn().Prepare(
		"insert into tbl_user_file(`user_name`,`file_hash`,`file_name`," +
			"`file_size`, `upload_at`) values (?,?,?,?,?)",
	)
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, fileHash, fileName, fileSize, time.Now())
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// QueryUserFileMetas 获取用户文件表元数据
func QueryUserFileMetas(username string, limit int) ([]UserFile, error) {
	stmt, err := mdb.DBConn().Prepare(
		"select file_hash, file_name, file_size," +
			"upload_at, last_update from tbl_user_file where user_name=? limit ?")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var userFiles []UserFile
	for rows.Next() {
		var userFile UserFile
		err = rows.Scan(&userFile.FileHash, &userFile.FileName, &userFile.FileSize, &userFile.UploadAt, &userFile.LastUpdated)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		userFiles = append(userFiles, userFile)
	}
	return userFiles, nil
}

// QueryUserFileMeta 获取用户单个文件信息
func QueryUserFileMeta(username string, filehash string) (*UserFile, error) {
	stmt, err := mdb.DBConn().Prepare(
		"select file_sha1,file_name,file_size,upload_at," +
			"last_update from tbl_user_file where user_name=? and file_sha1=?  limit 1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, filehash)
	if err != nil {
		return nil, err
	}

	ufile := UserFile{}
	if rows.Next() {
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize,
			&ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
	}

	return &ufile, nil
}
