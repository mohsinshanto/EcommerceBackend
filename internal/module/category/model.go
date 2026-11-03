package category

import "gorm.io/gorm"

type Category struct {
	gorm.Model

	Name     string `gorm:"size:100;not null;uniqueIndex"`
	ParentID *uint  `gorm:"index"`

	Parent   *Category   `gorm:"foreignKey:ParentID"`
	Children []*Category `gorm:"foreignKey:ParentID"`
}
