package model

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/spf13/viper"
	"github.com/Eraac/train-sniffer/utils"
)

var db *gorm.DB

func Init() {
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")
	hostname := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	name := viper.GetString("database.name")

	d, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, hostname, port, name))

	if err != nil {
		utils.Error(err.Error())
		os.Exit(utils.ErrorInitDatabase)
	}

	db = d

	db.SingularTable(true)

	registerModel()

	init_cache()
}

func registerModel() {
	db.AutoMigrate(&Station{}, &Train{}, &Passage{}, &Realtime{})
}

// Close db connection
func Close() {
	db.Close()
}
