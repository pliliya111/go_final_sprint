package handler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pliliya111/go_final_sprint/internal/database"
	"github.com/pliliya111/go_final_sprint/internal/middleware"
	"github.com/pliliya111/go_final_sprint/internal/model"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB // Глобальная переменная

func SetDB(database *sql.DB) {
	db = database
}

var (
	timeAdditionMS       = getEnvInt("TIME_ADDITION_MS", 1000)
	timeSubtractionMS    = getEnvInt("TIME_SUBTRACTION_MS", 1000)
	timeMultiplicationMS = getEnvInt("TIME_MULTIPLICATIONS_MS", 1000)
	timeDivisionMS       = getEnvInt("TIME_DIVISIONS_MS", 1000)
	validExpressionRegex = regexp.MustCompile(`^[\d\s\+\-\*\/\(\)]+$`)
)

func getEnvInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func parseExpression(expression string) []string {
	re := regexp.MustCompile(`\d+|\+|\-|\*|\/`)
	return re.FindAllString(expression, -1)
}

func AddExpression(c *gin.Context) {
	var request struct {
		Expression string `json:"expression"`
	}
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
		return
	}

	userID := userId.(int)
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid data"})
		return
	}

	if !validExpressionRegex.MatchString(request.Expression) {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Expression is not valid"})
		return
	}

	expressionID := uuid.New().String()
	expr := &model.Expression{
		ID:         expressionID,
		Expression: request.Expression,
		Status:     "pending",
		UserId:     userID,
	}

	if _, err := database.InsertExpression(c.Request.Context(), db, expr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save expression"})
		return
	}

	tokens := parseExpression(request.Expression)
	var tasks []*model.Task

	// Обработка операций умножения и деления
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "*" || tokens[i] == "/" {
			arg1ID := tokens[i-1]
			arg2ID := tokens[i+1]
			taskID := uuid.New().String()
			task := &model.Task{
				ID:           taskID,
				Arg1:         arg1ID,
				Arg2:         arg2ID,
				Operation:    tokens[i],
				ExpressionId: expressionID,
			}
			tasks = append(tasks, task)
			tokens[i-1] = taskID
			tokens[i] = ""
			tokens[i+1] = ""
		}
	}

	filteredTokens := []string{}
	for _, token := range tokens {
		if token != "" {
			filteredTokens = append(filteredTokens, token)
		}
	}
	tokens = filteredTokens

	// Обработка операций сложения и вычитания
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "+" || tokens[i] == "-" {
			arg1ID := tokens[i-1]
			arg2ID := tokens[i+1]
			taskID := uuid.New().String()
			task := &model.Task{
				ID:           taskID,
				Arg1:         arg1ID,
				Arg2:         arg2ID,
				Operation:    tokens[i],
				ExpressionId: expressionID,
			}
			tasks = append(tasks, task)
			tokens[i-1] = taskID
			tokens[i] = ""
			tokens[i+1] = ""
		}
	}

	// Вставляем все задачи одним запросом
	if len(tasks) > 0 {
		if err := database.InsertTasks(c.Request.Context(), db, tasks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save tasks"})
			return
		}

		expr.Result = tasks[len(tasks)-1].ID
	}
	expr.Status = "in_progress"

	if err := database.UpdateExpression(c.Request.Context(), db, expr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update expression"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": expressionID})
}

func GetExpressions(c *gin.Context) {
	ctx := c.Request.Context()

	expressions, err := database.GetExpressions(ctx, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch expressions"})
		return
	}

	expressionsList := make([]gin.H, 0, len(expressions))
	for _, expr := range expressions {
		expressionsList = append(expressionsList, gin.H{
			"id":     expr.ID,
			"status": expr.Status,
			"result": expr.Result,
		})
	}

	c.JSON(http.StatusOK, gin.H{"expressions": expressionsList})
}
func GetExpressionByID(c *gin.Context) {
	expressionID := c.Param("id")
	ctx := c.Request.Context()

	expr, err := database.GetExpressionByID(ctx, db, expressionID)
	if err != nil {
		if err.Error() == "expression not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "expression not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch expression"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expression": gin.H{
			"id":         expr.ID,
			"expression": expr.Expression,
			"status":     expr.Status,
			"result":     expr.Result,
		},
	})
}

func getOperationTime(operation string) int {
	switch operation {
	case "+":
		return timeAdditionMS
	case "-":
		return timeSubtractionMS
	case "*":
		return timeMultiplicationMS
	case "/":
		return timeDivisionMS
	default:
		return 0
	}
}

func GetTask(c *gin.Context) {
	ctx := c.Request.Context()

	task, err := database.GetNextPendingTask(ctx, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get task"})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no tasks available"})
		return
	}

	opTime := getOperationTime(task.Operation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task": gin.H{
			"id":             task.ID,
			"arg1":           task.Arg1,
			"arg2":           task.Arg2,
			"operation":      task.Operation,
			"operation_time": opTime,
			"expression_id":  task.ExpressionId,
		},
	})
}

func SubmitTaskResult(c *gin.Context) {
	var request struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid data"})
		return
	}

	ctx := c.Request.Context()
	if err := database.UpdateTaskResult(ctx, db, request.ID, request.Result); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to submit result"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "result submitted"})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func RegisterUser(c *gin.Context) {
	var request struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request data"})
		return
	}
	hashedPassword, err := hashPassword(request.Password)
	if err != nil {
		log.Fatalf("Ошибка хэширования пароля: %v", err)
	}
	user := model.User{
		Name:     request.Name,
		Password: hashedPassword,
	}

	userID, err := database.InsertUser(c.Request.Context(), db, &user)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to register user"})
		return
	}

	c.JSON(200, gin.H{
		"user_id": userID,
		"message": "User registered successfully",
	})
}

func LoginUser(c *gin.Context) {
	var request struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user, err := database.LoginUser(c.Request.Context(), db, request.Name, request.Password)
	if err != nil {
		switch err.Error() {
		case "user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "invalid password":
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}

	tokenString, err := middleware.GenerateToken(user.Name, int(user.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": user.ID,
		"name":    user.Name,
		"token":   tokenString,
	})
}
