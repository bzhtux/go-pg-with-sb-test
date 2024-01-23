package main

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bindings "github.com/bzhtux/servicebinding"
	"github.com/jinzhu/copier"
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
	// Debug output for servicebinding pkg
	b, err := bindings.NewBinding("postgres")
	if err != nil {
		log.Printf("Error while getting bindings: %s\n", err.Error())
	} else {
		err := filepath.Walk("/bindings", func(bpath string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				log.Printf(" -> %v id dir", bpath)
			} else {
				if !info.IsDir() {
					log.Printf(" -> %v is file", bpath)
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("walk error [%v]\n", err)
		}
	}

	copier.Copy(&cfg.Database, &b)

	fmt.Printf("DB config: %v", &cfg.Database)

	// Setup new DB connection
	var dsn = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.Username, cfg.Database.Database, cfg.Database.Password)
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// log.Printf("*** Error connectinng to DB: %s", err.Error())
		log.Fatal(err.Error())
	}

	fmt.Printf("DBConn: %v", conn)
	conn.AutoMigrate(&Books{})

	// New DB entry
	newbook := Books{
		Title:  "The Hitchhiker's Guide to the Galaxy",
		Author: "Douglas Adams",
	}

	// Ensure entry doesn't exist yet
	entry := conn.Where("Title = ?", newbook.Title)
	if entry.RowsAffected != 0 {
		log.Printf("Book %s already exists in DB", newbook.Title)
	} else {
		conn.Create(&newbook)
		log.Printf("New book %s by %s has been recorded in DB", newbook.Title, newbook.Author)
	}
}
