package db

import (
	"database/sql"
	"fmt"
	"log"
	mdb "objectstore-server/db/mysql"
)

// OnFileUploadFinished 文件上传成功，保存元数据
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	stmt, err := mdb.DBConn().Prepare(
		"insert ignore into tbl_file(`file_hash`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)")
	if err != nil {
		log.Println("Failed to prepare statement,", err)
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		log.Println(err)
		return false
	}

	if rf, err := ret.RowsAffected(); err == nil {
		if rf <= 0 {
			// 之前插入过相同的文件
			log.Println("This file", filehash, "has been uploaded")
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
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFileDTO{}
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileName, &tfile.FileSize, &tfile.FileAddr)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &tfile, nil
}

// UpdateFileLocation 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mdb.DBConn().Prepare(
		"update tbl_file set`file_addr`=? where  `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
			return false
		}
		return true
	}
	return false
}
