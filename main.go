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

type Book struct {
	*gorm.Model
	ID     uint   `gorm:"index;unique;autoIncrement"`
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

	if err := conn.AutoMigrate(&Book{}); err != nil {
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
		newbook := Book{
			Title:  "The Hitchhiker's Guide to the Galaxy",
			Author: "Douglas Adams",
		}

		// Ensure entry doesn't exist yet
		entry := conn.Where("Title = ?", newbook.Title).First(&newbook)
		log.Printf("entry.Error: %v", entry.Error)
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

	router.GET("/add2", func(c *gin.Context) {
		// New DB entry
		newbook2 := Book{
			Title:  "Alice's Adventures in Wonderland",
			Author: "Lewis Carroll",
		}

		// Ensure entry doesn't exist yet
		entry2 := conn.Where("Title = ?", newbook2.Title).First(&newbook2)
		log.Printf("entry.Error: %v", entry2.Error)
		if entry2.RowsAffected != 0 {
			log.Printf("Book %s already exists in DB", newbook2.Title)
			c.JSON(http.StatusConflict, gin.H{
				"status": "Conflict",
				"data":   "Entry already exists in DB",
			})
		} else {
			conn.Create(&newbook2)
			log.Printf("New book %s by %s has been recorded in DB", newbook2.Title, newbook2.Author)
			c.JSON(http.StatusOK, gin.H{
				"status": "OK",
				"data":   "New book recorded",
			})
		}
	})

	router.GET("/list", func(c *gin.Context) {
		var allbooks []Book
		r := conn.Find(&allbooks)
		if r.RowsAffected == 0 {

			c.JSON(http.StatusNotFound, gin.H{
				"status": "Not Found",
				"data":   "No record found in DB",
			})
			log.Printf("No entry were found in DB.")
		} else {

			c.JSON(http.StatusOK, gin.H{
				"data": allbooks,
			})
			log.Printf("%d record(s) found in DB", r.RowsAffected)
		}
	})

	router.GET("/clean", func(c *gin.Context) {
		// r := conn.Where("1 = 1").Delete(&Book{})
		r := conn.Exec("DELETE FROM books")
		if r.RowsAffected == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "Internal server Error",
				"data":   "Something went wrong when deleting All Books",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status": "OK",
				"data":   "All records have been deleted",
			})
		}
	})

	router.GET("/clean/:bookid", func(c *gin.Context) {
		delbookID := c.Params.ByName("bookid")
		var book = Book{}
		r3 := conn.Where("ID = ?", delbookID).First(&book)
		if r3.RowsAffected == 0 {
			log.Printf("No book found with ID %v", delbookID)
			c.JSON(http.StatusNotFound, gin.H{
				"status": "Not Found",
				"data":   "No book with ID " + delbookID + " was found in DB",
			})
		} else {
			log.Printf("Deleting book: %v", book.Title)
			// del := conn.Where("ID = ?", delbookID).Delete(&Book{})
			del := conn.Unscoped().Where("ID = ?", delbookID).Delete(&Book{})
			if del.RowsAffected == 0 {
				log.Printf("Book with ID %v was not deleted from DB", delbookID)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status": "Internal Server Error",
					"data":   "Something went wrong when deleting Book with ID " + delbookID,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status": "OK",
					"data":   "Book with ID " + delbookID + " was successfully deleted",
				})
			}
		}
	})

	router.Run(":8080")
}
