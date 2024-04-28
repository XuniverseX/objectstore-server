package db

import (
	"database/sql"
	"fmt"
	mdb "objectstore-server/db/mysql"
)

// OnFileUploadFinished 文件上传成功，保存元数据
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	stmt, err := mdb.DBConn().Prepare(
		"insert ignore into tbl_file(`file_hash`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)")
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

type TableFileDTO struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// GetFileMeta 从mysql获取元数据
func GetFileMeta(filehash string) (*TableFileDTO, error) {
	stmt, err := mdb.DBConn().Prepare("select file_hash, file_name, file_size, file_addr" +
		" from tbl_file where file_hash=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	tfile := TableFileDTO{}
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileName, &tfile.FileSize, &tfile.FileAddr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer stmt.Close()

	return &tfile, nil
}
