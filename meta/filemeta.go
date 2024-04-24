package meta

import mysqlDb "objectstore-server/db"

// FileMeta 文件元信息结构
type FileMeta struct {
	Hash       string
	FileName   string
	FileSize   int64
	Location   string
	UploadTime string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta 新增/更新文件元信息
func UpdateFileMeta(meta FileMeta) {
	fileMetas[meta.Hash] = meta
}

// UpdateFileMetaToDB 新增/更新文件元信息到数据库
func UpdateFileMetaToDB(meta FileMeta) bool {
	return mysqlDb.OnFileUploadFinished(meta.Hash, meta.FileName, meta.FileSize, meta.Location)
}

// GetFileMeta 通过hash获取文件的元信息对象
func GetFileMeta(hash string) FileMeta {
	return fileMetas[hash]
}

// GetFileMetaFromDB 从mysql获取文件元信息
func GetFileMetaFromDB(hash string) (FileMeta, error) {
	tfile, err := mysqlDb.GetFileMeta(hash)
	if err != nil {
		return FileMeta{}, err
	}

	fmeta := FileMeta{
		Hash:     tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return fmeta, nil
}

// RemoveFileMeta 删除文件元信息
func RemoveFileMeta(hash string) {
	delete(fileMetas, hash)
}
