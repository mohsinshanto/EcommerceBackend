package category

import "gorm.io/gorm"

type Category struct {
	gorm.Model

	Name     string      `gorm:"size:100;not null;uniqueIndex" json:"name"`
	ParentID *uint       `gorm:"index" json:"parent_id"`
	Parent   *Category   `gorm:"foreignKey:ParentID" json:"parent"`
	Children []*Category `gorm:"foreignKey:ParentID" json:"children"`
}
