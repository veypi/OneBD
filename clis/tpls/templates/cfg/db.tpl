
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

var db *gorm.DB

var cmdDB = CMD.SubCommand("db", "database operations")
var cmdMigrate = cmdDB.SubCommand("migrate", "migrate database")
var ObjList = make([]any, 0, 10)

func init() {
	cmdMigrate.Command = func() error {
		return DB().AutoMigrate(ObjList...)
	}
	cmdDB.SubCommand("drop", "drop database").Command = func() error {
		return DB().Migrator().DropTable(ObjList...)
	}
}

func DB() *gorm.DB {
	if db == nil {
		var err error
		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN: Config.DSN,
		}), &gorm.Config{})
		if err != nil {
			panic(err)
		}
	}
	return db
}
