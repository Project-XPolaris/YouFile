package database

import "gorm.io/gorm"

type Thumbnail struct {
	gorm.Model
	Path     string
	Checksum string
}
