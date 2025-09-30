package topic

import (
	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/taskengine"
	"github.com/xxcheng123/cloudpan189-share/internal/types/autoingest"
)

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

type AutoIngestRefreshSubscribeRequest struct {
	PlanId     int64                 `json:"planId"`
	ParentPath string                `json:"parentPath"`
	OnConflict autoingest.OnConflict `json:"onConflict"`
	Offset     int64                 `json:"offset"` // 如果是 0 表示全部加一遍
	UpUserId   string                `json:"upUserId"`
	CloudToken int64                 `json:"cloudToken"`
}

func (r AutoIngestRefreshSubscribeRequest) Topic() taskengine.Topic {
	return taskengine.Topic(KeyAutoIngestRefreshSubscribe)
}
