package models

import (
	"time"

	"github.com/xxcheng123/cloudpan189-share/internal/pkgs/datatypes"
)

const (
	SourceType = "subscribe"
)

const (
	OnConflictRename  = "rename"
	OnConflictAbandon = "abandon"
)

type AutoIngestPlan struct {
	ID                 int64             `gorm:"primaryKey" json:"id"`
	Name               string            `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Enabled            bool              `gorm:"column:enabled;type:tinyint(1);default:1" json:"enabled"`
	AutoIngestInterval int64             `gorm:"column:auto_ingest_interval;type:bigint(20);default:30" json:"autoIngestInterval"` // 单位分钟
	SourceType         string            `gorm:"column:source_type" json:"sourceType"`
	Offset             int64             `gorm:"column:offset;type:bigint(20);default:0;" json:"offset"`                      // 偏移量
	ParentPath         string            `gorm:"column:parent_path;type:varchar(255);not null;default:'';" json:"parentPath"` // 父目录路径
	OnConflict         string            `gorm:"column:on_conflict;type:varchar(255);not null;default:'';" json:"onConflict"` // 冲突处理策略
	AddCount           int64             `gorm:"column:add_count;type:bigint(20);default:0;" json:"addCount"`                 // 新增挂载数
	FailedCount        int64             `gorm:"column:failed_count;type:bigint(20);default:0;" json:"failedCount"`           // 失败挂载数
	Addition           datatypes.JSONMap `gorm:"column:addition;type:json;default:'';" json:"addition"`
	CreatedAt          time.Time         `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt          time.Time         `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (a *AutoIngestPlan) TableName() string {
	return "auto_ingest_plans"
}

type AutoIngestLog struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	PlanId    int64     `gorm:"column:plan_id;type:bigint(20);not null;default:0;" json:"planId"`
	Level     string    `gorm:"column:level;type:varchar(64);not null;default:info" json:"level"`
	Content   string    `gorm:"column:content;type:text;not null;default:'';" json:"content"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (a *AutoIngestLog) TableName() string {
	return "auto_ingest_logs"
}

type AutoIngestPlanSubscribeAddition struct {
	UpUserId string `json:"upUserId"`
}

func (a *AutoIngestPlanSubscribeAddition) JSONMap() datatypes.JSONMap {
	m, _ := datatypes.FromStruct(a)

	return m
}
