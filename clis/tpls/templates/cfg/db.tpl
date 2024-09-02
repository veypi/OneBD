
//
// db.go
// Copyright (C) 2024 veypi <i@veypi.com>
// {{.common.date}}
// Distributed under terms of the MIT license.
//

package cfg

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() error {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: Config.DSN,
	}), &gorm.Config{})
	DB = db
	return err
}
