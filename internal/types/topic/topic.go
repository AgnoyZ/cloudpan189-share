package topic

import "github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"

type Request interface {
	Topic() taskengine.Topic
}

type FileScanFileRequest struct {
	FileId int64 `json:"fileId"` // 入口地址
	Deep   bool  `json:"deep"`   // 是否深度扫描
}

func (r FileScanFileRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyFileScanFile)
}

type FileClearFileRequest struct {
	FileId int64 `json:"fileId"` // 入口地址
}

func (r FileClearFileRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyFileClearFile)
}
