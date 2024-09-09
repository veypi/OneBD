//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package {{.common.model}}

import (
    "{{.common.repo}}/cfg"
	"gorm.io/gorm"
	"time"
)

var Mirgate = cfg.CMD.SubCommand("migrate", "migrate database")

func init() {
	Mirgate.Command = AutoMigrate
}


type BaseModel struct {
    ID        string         `json:"id" gorm:"primaryKey;type:varchar(32)" methods:"get,put,patch,list,delete" parse:"path"`
    CreatedAt time.Time      `json:"created_at" methods:"get,list" parse:"query"`
	UpdatedAt time.Time      `json:"updated_at" methods:"get,list" parse:"query"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func AutoMigrate() error {
	err := cfg.ConnectDB()
	if err != nil {
		return err
	}
	db := cfg.DB

	return nil
}
