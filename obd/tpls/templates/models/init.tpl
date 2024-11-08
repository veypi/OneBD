//
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package {{.common.model}}

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strings"
	"time"
)

type BaseModel struct {
  // ID        uint           `json:"id" gorm:"primaryKey" methods:"get,patch,delete" parse:"path"`
  ID        string         `json:"id" gorm:"primaryKey;type:varchar(32)" methods:"get,put,patch,delete" parse:"path"`
  CreatedAt time.Time      `json:"created_at" methods:"*list" parse:"query"`
	UpdatedAt time.Time      `json:"updated_at" methods:"*list" parse:"query"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = strings.ReplaceAll(uuid.New().String(), "-", "")
	}
	return nil
}
