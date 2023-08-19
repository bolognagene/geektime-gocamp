package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	err := db.AutoMigrate(&User{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&UserProfile{})
	if err != nil {
		panic(err)
	}

	return err
}
