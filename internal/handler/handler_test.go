package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pliliya111/go_final_sprint/internal/database"
	"github.com/pliliya111/go_final_sprint/internal/handler"
	"github.com/pliliya111/go_final_sprint/internal/middleware"
	"github.com/pliliya111/go_final_sprint/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	db *sql.DB
)

func setupDatabase() {
	var err error
	projectDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	tempDBFile := filepath.Join(projectDir, "test_store.db")
	db, err = database.OpenDatabase(tempDBFile)
	if err != nil {
		panic(err)
	}

	if err = database.CreateTables(context.TODO(), db); err != nil {
		panic(err)
	}

	handler.SetDB(db)
}

func teardownDatabase() {
	if db != nil {
		db.Close()
	}
	projectDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tempDBFile := filepath.Join(projectDir, "test_store.db")
	os.Remove(tempDBFile)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/v1/register", handler.RegisterUser)
	r.POST("/api/v1/login", handler.LoginUser)

	auth := r.Group("/api/v1")
	auth.Use(middleware.AuthMiddleware())

	auth.POST("/calculate", handler.AddExpression)
	auth.GET("/expressions", handler.GetExpressions)
	auth.GET("/expressions/:id", handler.GetExpressionByID)

	r.GET("/internal/task", handler.GetTask)
	r.POST("/internal/task", handler.SubmitTaskResult)

	return r
}

func TestMain(m *testing.M) {
	setupDatabase()
	code := m.Run()
	teardownDatabase()
	os.Exit(code)
}

func TestAddExpression(t *testing.T) {
	router := setupRouter()

	user := &model.User{
		Name:     "user_1",
		Password: "password",
	}

	userId, err := database.InsertUser(context.Background(), db, user)
	assert.NoError(t, err)

	token, err := middleware.GenerateToken("user_1", int(userId))
	assert.NoError(t, err)

	payload := `{"expression": "2 + 3 * 4"}`
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expressionID := response["id"]
	assert.NotEmpty(t, expressionID)

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM expressions WHERE id = ?", expressionID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGetTask(t *testing.T) {
	router := setupRouter()

	payload := `{"expression": "2 + 3 * 4"}`
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	req, _ = http.NewRequest("GET", "/internal/task", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var taskResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &taskResponse)
	assert.NoError(t, err)

	task := taskResponse["task"].(map[string]interface{})
	assert.NotEmpty(t, task["id"])
	assert.Equal(t, "*", task["operation"])
}

func TestSubmitTaskResult(t *testing.T) {
	router := setupRouter()

	payload := `{"expression": "2 + 3 * 4"}`
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	req, _ = http.NewRequest("GET", "/internal/task", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var taskResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &taskResponse)
	assert.NoError(t, err)

	task := taskResponse["task"].(map[string]interface{})
	taskID := task["id"].(string)

	resultPayload := `{"id": "` + taskID + `", "result": 12}`
	req, _ = http.NewRequest("POST", "/internal/task", bytes.NewBufferString(resultPayload))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result float64
	err = db.QueryRow("SELECT result FROM tasks WHERE id = ?", taskID).Scan(&result)
	assert.NoError(t, err)
	assert.Equal(t, 12.0, result)
}

func insertTestExpression(userId int, expression string) (string, error) {
	expr := &model.Expression{
		ID:         uuid.New().String(),
		UserId:     userId,
		Expression: expression,
		Status:     "in_progress",
	}

	expressionID, err := database.InsertExpression(context.Background(), db, expr)
	if err != nil {
		return "", err
	}
	return expressionID, nil
}

func TestGetExpressions(t *testing.T) {
	router := setupRouter()

	user := &model.User{
		Name:     "user_2",
		Password: "password",
	}

	userId, err := database.InsertUser(context.Background(), db, user)
	assert.NoError(t, err)

	token, err := middleware.GenerateToken(user.Name, int(userId))
	assert.NoError(t, err)

	_, err = insertTestExpression(int(userId), "2 + 3 * 4")
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/expressions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expressionsList, ok := response["expressions"].([]interface{})
	assert.True(t, ok)
	assert.Greater(t, len(expressionsList), 0)

	for _, expr := range expressionsList {
		exprMap, ok := expr.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, exprMap, "id")
		assert.Contains(t, exprMap, "status")
		assert.Contains(t, exprMap, "result")
	}
}

func TestGetExpressionByID(t *testing.T) {
	router := setupRouter()

	user := &model.User{
		Name:     "user_3",
		Password: "password",
	}

	userId, err := database.InsertUser(context.Background(), db, user)
	assert.NoError(t, err)

	token, err := middleware.GenerateToken(user.Name, int(userId))
	assert.NoError(t, err)

	expressionID, err := insertTestExpression(int(userId), "5 + 7")
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/expressions/"+expressionID, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	exprMap, ok := response["expression"].(map[string]interface{})
	assert.True(t, ok)

	assert.Contains(t, exprMap, "id")
	assert.Contains(t, exprMap, "status")
	assert.Contains(t, exprMap, "result")
	assert.Equal(t, expressionID, exprMap["id"])
	assert.Equal(t, "in_progress", exprMap["status"])
}
