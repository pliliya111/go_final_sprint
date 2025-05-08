package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pliliya111/go_final_sprint/internal/handler"
	"github.com/pliliya111/go_final_sprint/internal/model"
	"github.com/stretchr/testify/assert"
)

var (
	expressions = make(map[string]*model.Expression) // Хранение выражений
	tasks       []*model.Task                        // Очередь задач
	taskMap     = make(map[string]*model.Task)       // Мапа для быстрого поиска задач по ID
	results     = make(map[string]interface{})       // Результаты задач
	mutex       = &sync.Mutex{}                      // Мьютекс для потокобезопасности
)

func resetState() {
	handler.Expressions = make(map[string]*model.Expression)
	handler.Tasks = []*model.Task{}
	handler.TaskMap = make(map[string]*model.Task)
	handler.Results = make(map[string]interface{})
	handler.Mutex = &sync.Mutex{}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/api/v1/calculate", handler.AddExpression)
	r.GET("/api/v1/expressions", handler.GetExpressions)
	r.GET("/api/v1/expressions/:id", handler.GetExpressionByID)
	r.GET("/internal/task", handler.GetTask)
	r.POST("/internal/task", handler.SubmitTaskResult)
	return r
}

func TestAddExpression(t *testing.T) {
	router := setupRouter()
	resetState()
	payload := `{"expression": "2 + 3 * 4"}`
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expressionID := response["id"]
	assert.NotEmpty(t, expressionID)

	assert.Equal(t, 2, len(handler.Tasks))
	assert.Equal(t, 2, len(handler.TaskMap))
}

func TestGetTask(t *testing.T) {
	router := setupRouter()
	resetState()

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
	resetState()

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

	mutex.Lock()
	result, exists := handler.Results[taskID]
	mutex.Unlock()

	assert.True(t, exists)
	assert.Equal(t, 12.0, result)
}

func TestGetExpressions(t *testing.T) {
	router := setupRouter()
	resetState()
	payload := `{"expression": "2 + 3 * 4"}`
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	req, _ = http.NewRequest("GET", "/api/v1/expressions", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expressionsList := response["expressions"].([]interface{})
	assert.Equal(t, 1, len(expressionsList))
}

func TestGetExpressionByID(t *testing.T) {
	router := setupRouter()
	resetState()
	payload := `{"expression": "2 + 3 * 4"}`
	req, _ := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	expressionID := response["id"]

	req, _ = http.NewRequest("GET", "/api/v1/expressions/"+expressionID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var exprResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &exprResponse)
	assert.NoError(t, err)

	expr := exprResponse["expression"].(map[string]interface{})
	assert.Equal(t, expressionID, expr["id"])
	assert.Equal(t, "in_progress", expr["status"])
}
