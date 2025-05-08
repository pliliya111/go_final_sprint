package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pliliya111/go_final_sprint/internal/handler"
	"github.com/pliliya111/go_final_sprint/internal/middleware"
)

func createTables(ctx context.Context, db *sql.DB) error {
	fmt.Println("создаем таблицы")
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT,
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		Result TEXT,
		user_id INTEGER NOT NULL,
	
		FOREIGN KEY (user_id)  REFERENCES expressions (id)
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		return err
	}

	return nil
}
func main() {
	ctx := context.TODO()

	db, err := sql.Open("sqlite3", "store.db")

	if err != nil {
		panic(err)
	}
	handler.SetDB(db)
	defer db.Close()

	err = db.PingContext(ctx)
	if err != nil {
		panic(err)
	}

	if err = createTables(ctx, db); err != nil {
		panic(err)
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
