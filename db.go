package wilhelmiina

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDatabase(databaseName string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(databaseName))
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Migrates datatypes to database, should be run only on the first run of the app
func CreateTables(db *gorm.DB) error {
	err := db.AutoMigrate(&User{}, &Course{}, &Group{}, &GroupReservation{}, &GroupTime{}, &Message{}, &MessageReciever{}, Subject{})
	return err
}
