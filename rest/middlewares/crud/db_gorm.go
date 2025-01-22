//
// gorm.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-11-29 17:42
// Distributed under terms of the GPL license.
//

package crud

import (
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, graph *StructGraph) error {
	items := make([]any, 0, 10)
	for _, obj := range graph.Nodes {
		items = append(items, obj.v.Interface())
	}
	db.DisableForeignKeyConstraintWhenMigrating = true
	err := db.AutoMigrate(items...)
	if err != nil {
		return err
	}
	db.DisableForeignKeyConstraintWhenMigrating = false
	return db.AutoMigrate(items...)
}

func AutoDrop(db *gorm.DB, graph *StructGraph) error {
	items := make([]any, 0, 10)
	for _, obj := range graph.Nodes {
		items = append(items, obj.v.Interface())
	}
	return db.Migrator().DropTable(items...)
}
