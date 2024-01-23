package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	bindings "github.com/bzhtux/servicebinding"
	"github.com/gin-gonic/gin"
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
	ID     uint   `gorm:"index;unique"`
	Title  string `gorm:"unique;primaryKey" json:"title" binding:"required"`
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
		log.Printf(err.Error())
	}

	fmt.Printf("DBConn: %v", conn)

	if err := conn.AutoMigrate(&Books{}); err != nil {
		log.Printf("Error Migrating DB: %v", err.Error())
	}

	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "OK",
			"data":   "Hello TAP !",
		})
	})

	router.GET("/add", func(c *gin.Context) {
		// New DB entry
		newbook := Books{
			Title:  "The Hitchhiker's Guide to the Galaxy",
			Author: "Douglas Adams",
		}

		// Ensure entry doesn't exist yet
		entry := conn.Where("Title = ?", newbook.Title)
		if entry.RowsAffected != 0 {
			log.Printf("Book %s already exists in DB", newbook.Title)
			c.JSON(http.StatusConflict, gin.H{
				"status": "Conflict",
				"data":   "Entry already exists in DB",
			})
		} else {
			conn.Create(&newbook)
			log.Printf("New book %s by %s has been recorded in DB", newbook.Title, newbook.Author)
			c.JSON(http.StatusOK, gin.H{
				"status": "OK",
				"data":   "New book recorded",
			})
		}
	})

	router.GET("/list", func(c *gin.Context) {
		var allbooks []Books
		entries := conn.Find(&allbooks)
		if entries.RowsAffected == 0 {
			log.Printf("No entry were found in DB.")
			c.JSON(http.StatusNotFound, gin.H{
				"status": "Not Found",
				"data":   "No record found in DB",
			})
		} else {
			log.Printf("%d record(s) found in DB", entries.RowsAffected)
			c.JSON(http.StatusOK, gin.H{
				"data": allbooks,
			})
		}
	})

	router.GET("/clean", func(c *gin.Context) {
		// conn.Exec("DELETE FROM books")
		r := conn.Where("1 = 1").Delete(&Books{})
		if r.RowsAffected == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Internal server Error",
				"data":   "Something went wrong when deleting All Books",
			})
		}
	})

	router.GET("/clean/:id", func(c *gin.Context) {
		bookID := c.Params.ByName("id")
		var book = Books{}
		r := conn.Where("ID = ?", bookID).First(&book)
		if r.RowsAffected == 0 {
			log.Printf("No book found with ID %v", bookID)
			c.JSON(http.StatusNotFound, gin.H{
				"status": "Not Found",
				"data":   "No book with ID " + bookID + " was found in DB",
			})
		} else {
			r := conn.Delete(&book, bookID)
			if r.RowsAffected == 0 {
				log.Printf("Book with ID %v was not deleted from DB", bookID)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status": "Internal Server Error",
					"data":   "Something went wrong when deleting Book with ID " + bookID,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": "OK",
					"data":   "Book with ID " + bookID + " was successfully deleted",
				})
			}
		}
	})

	router.Run(":8080")
}
