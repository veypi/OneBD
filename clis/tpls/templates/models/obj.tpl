//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package {{.package}}

import (
	"gorm.io/gorm"
	"time"
)

type {{.Obj}} struct {
    CreatedAt time.Time      `json:"created_at" methods:"get,post,put,patch,list,delete"`
	UpdatedAt time.Time      `json:"updated_at" methods:"get,post,patch"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
