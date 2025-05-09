package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pliliya111/go_final_sprint/internal/database"
	"github.com/pliliya111/go_final_sprint/internal/handler"
	"github.com/pliliya111/go_final_sprint/internal/middleware"
)

func main() {
	ctx := context.TODO()

	db, err := database.OpenDatabase("store.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	handler.SetDB(db)

	err = db.PingContext(ctx)
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err = database.CreateTables(ctx, db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
	r := gin.Default()

	r.POST("/api/v1/register", handler.RegisterUser)
	r.POST("/api/v1/login", handler.LoginUser)

	auth := r.Group("/api/v1")
	auth.Use(middleware.AuthMiddleware())

	// Эти маршруты требуют авторизации
	auth.POST("/calculate", handler.AddExpression)
	auth.GET("/expressions", handler.GetExpressions)
	auth.GET("/expressions/:id", handler.GetExpressionByID)

	// Остальные маршруты
	r.GET("/internal/task", handler.GetTask)
	r.POST("/internal/task", handler.SubmitTaskResult)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
