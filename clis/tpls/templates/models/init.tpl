//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package {{.common.model}}

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
    ID        string         `json:"id" gorm:"primaryKey;type:varchar(32)" methods:"get,post,put,patch,list,delete" parse:"path"`
    CreatedAt time.Time      `json:"created_at" methods:"get,list" parse:"query"`
	UpdatedAt time.Time      `json:"updated_at" methods:"get,list" parse:"query"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
