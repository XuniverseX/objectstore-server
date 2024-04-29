package db

import (
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
func OnUserFileUploadFinished(username, fileHash, fileName string, fileSize int64) (bool, error) {
	stmt, err := mdb.DBConn().Prepare(
		"insert into tbl_user_file(`user_name`,`file_hash`,`file_name`," +
			"`file_size`, `upload_at`) values (?,?,?,?,?)",
	)
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, fileHash, fileName, fileSize, time.Now())
	if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
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
