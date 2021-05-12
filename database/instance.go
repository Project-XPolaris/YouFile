package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Instance *gorm.DB

func ConnectToDatabase() error {
	var err error
	Instance, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	err = Instance.AutoMigrate(&Thumbnail{})
	if err != nil {
		return err
	}
	return nil
}
