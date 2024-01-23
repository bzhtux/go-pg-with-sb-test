package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bindings "github.com/bzhtux/servicebinding"
	"github.com/jinzhu/copier"
	// "github.com/bzhtux/servicebinding/bindings"
)

type Config struct {
	Dir struct {
		Root string
	}
	App struct {
		Name    string
		Desc    string
		Version string
		Port    int
	}
	Database struct {
		Host     string
		Port     int
		Username string
		Password string
		Database string
		Type     string
		SSL      bool
	}
}

type Books struct {
	*gorm.Model
	ID     uint   `gorm:"primaryKey"`
	Title  string `gorm:"not null" json:"title" binding:"required"`
	Author string `gorm:"not null" json:"author" binding:"required"`
}

func main() {
	cfg := Config{}
	b, err := bindings.NewBinding("postgres")
	if err != nil {
		log.Printf("Error while getting bindings: %s\n", err.Error())
	}
	copier.Copy(&cfg.Database, &b)

	fmt.Printf("DB config: %v", &cfg.Database)

	var dsn = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Username, cfg.Database.Database, cfg.Database.Password)
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// log.Printf("*** Error connectinng to DB: %s", err.Error())
		log.Fatal(err.Error())
	}

	fmt.Printf("DBConn: %v", conn)
	// conn.Automigrate(B&ooks{})
}
