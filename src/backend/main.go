package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Todo struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

func main() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Todo{})

	r := gin.Default()

	// Enhanced CORS configuration
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// GET all todos
	r.GET("/todos", func(c *gin.Context) {
		var todos []Todo
		result := db.Find(&todos)
		if result.Error != nil {
			c.JSON(500, gin.H{
				"error":   result.Error.Error(),
				"message": "Failed to fetch todos",
			})
			return
		}
		c.JSON(200, todos)
	})

	r.POST("/todos", func(c *gin.Context) {
		var todo Todo
		if err := c.BindJSON(&todo); err != nil {
			c.JSON(400, gin.H{
				"error":   err.Error(),
				"message": "Invalid request body",
			})
			return
		}

		result := db.Create(&todo)
		if result.Error != nil {
			c.JSON(500, gin.H{
				"error":   result.Error.Error(),
				"message": "Failed to create todo",
			})
			return
		}

		c.JSON(200, todo)
	})

	r.PUT("/todos/:id", func(c *gin.Context) {
		var todo Todo
		id := c.Param("id")
		if err := db.First(&todo, id).Error; err != nil {
			c.JSON(404, gin.H{"error": "Todo not found"})
			return
		}
		if err := c.BindJSON(&todo); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		db.Save(&todo)
		c.JSON(200, todo)
	})

	r.DELETE("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")
		var todo Todo
		if err := db.First(&todo, id).Error; err != nil {
			c.JSON(404, gin.H{"error": "Todo not found"})
			return
		}
		db.Delete(&todo)
		c.JSON(200, gin.H{"message": "Todo deleted"})
	})

	r.Run(":8080")
}
