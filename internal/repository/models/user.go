package models

import (
	"time"
)

type User struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"column:username;type:varchar(255);not null;uniqueIndex" json:"username"`
	Password  string    `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Status    int8      `gorm:"column:status;type:tinyint(1);default:1" json:"status"`
	IsAdmin   bool      `gorm:"column:is_admin;type:tinyint(1);default:0" json:"isAdmin"`
	GroupID   int64     `gorm:"column:group_id;type:bigint(20);default:0" json:"groupId"`
	Version   int       `gorm:"column:version;type:int(11);default:1" json:"version"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) Valid() bool {
	return u.Status == 1
}

type UserGroup struct {
	ID   int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"column:name;type:varchar(255);not null;uniqueIndex" json:"name"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (u *UserGroup) TableName() string {
	return "user_groups"
}

type Group2File struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PType     string    `gorm:"size:512;uniqueIndex:unique_index;default:'p';column:ptype" json:"-"`
	GroupId   string    `gorm:"size:512;uniqueIndex:unique_index" json:"groupId"`
	FileId    string    `gorm:"size:512;uniqueIndex:unique_index" json:"fileId"`
	V2        string    `gorm:"size:512;uniqueIndex:unique_index" json:"-"`
	V3        string    `gorm:"size:512;uniqueIndex:unique_index" json:"-"`
	V4        string    `gorm:"size:512;uniqueIndex:unique_index" json:"-"`
	V5        string    `gorm:"size:512;uniqueIndex:unique_index" json:"-"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime;type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime;type:datetime;default:CURRENT_TIMESTAMP;on update:CURRENT_TIMESTAMP" json:"updatedAt"`
}

func (g *Group2File) TableName() string {
	return "group2files"
}
