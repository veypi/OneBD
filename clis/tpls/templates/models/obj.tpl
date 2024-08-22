//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package models

import (
	"gorm.io/gorm"
	"time"
)

type {{.obj}} struct {
    CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

}
