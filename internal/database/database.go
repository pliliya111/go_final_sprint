package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/pliliya111/go_final_sprint/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func CreateTables(ctx context.Context, db *sql.DB) error {
	fmt.Println("Creating tables")
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT, 
		name TEXT,
		password TEXT
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id TEXT PRIMARY KEY, 
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		Result TEXT,
		user_id INTEGER NOT NULL,
		FOREIGN KEY (user_id)  REFERENCES users (id)
	);`
		tasksTable = `
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			arg1 TEXT,
			arg2 TEXT,
			operation TEXT NOT NULL,
			result TEXT,
			expression_id TEXT NOT NULL,
			status TEXT, 
			FOREIGN KEY (expression_id) REFERENCES expressions (id)
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		log.Printf("Error creating users table: %v", err)
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		log.Printf("Error creating expressions table: %v", err)
		return err
	}

	if _, err := db.ExecContext(ctx, tasksTable); err != nil {
		log.Printf("Error creating tasks table: %v", err)
		return err
	}

	return nil
}

func OpenDatabase(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func isUserExists(ctx context.Context, db *sql.DB, username string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE name = $1", username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func InsertUser(ctx context.Context, db *sql.DB, user *model.User) (int64, error) {
	exists, err := isUserExists(ctx, db, user.Name)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, errors.New("username already exist")
	}

	var q = `INSERT INTO users (name, password) VALUES ($1, $2)`
	result, err := db.ExecContext(ctx, q, user.Name, user.Password)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	return result.LastInsertId()
}

func LoginUser(ctx context.Context, db *sql.DB, name, password string) (*model.User, error) {
	user := &model.User{}

	err := db.QueryRowContext(
		ctx,
		"SELECT id, name, password FROM users WHERE name = $1",
		name,
	).Scan(&user.ID, &user.Name, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if !checkPasswordHash(password, user.Password) {
		return nil, fmt.Errorf("")
	}

	return &model.User{
		ID:   user.ID,
		Name: user.Name,
	}, nil
}

func InsertExpression(ctx context.Context, db *sql.DB, expr *model.Expression) (string, error) {
	var q = `INSERT INTO expressions (id, expression, status, result, user_id) VALUES ($1, $2, $3, $4, $5)`
	_, err := db.ExecContext(ctx, q, expr.ID, expr.Expression, expr.Status, expr.Result, expr.UserId)
	if err != nil {
		fmt.Println(err)
		return "", fmt.Errorf("failed to insert expression: %w", err)
	}

	return expr.ID, nil
}

func UpdateExpression(ctx context.Context, db *sql.DB, expr *model.Expression) error {
	var q = `UPDATE expressions SET expression = $1, status = $2, result = $3 WHERE id = $4`
	result, err := db.ExecContext(ctx, q, expr.Expression, expr.Status, expr.Result, expr.ID)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to update expression: %w", err)
	}
	fmt.Println(expr.ID)
	fmt.Println(result.RowsAffected())
	return nil
}

func InsertTasks(ctx context.Context, db *sql.DB, tasks []*model.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	var q strings.Builder
	q.WriteString("INSERT INTO tasks (id, arg1, arg2, operation, expression_id) VALUES ")

	args := make([]interface{}, 0, len(tasks)*5)
	for i, task := range tasks {
		if i > 0 {
			q.WriteString(", ")
		}
		q.WriteString(fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
			i*5+1, i*5+2, i*5+3, i*5+4, i*5+5))

		args = append(args, task.ID, task.Arg1, task.Arg2, task.Operation, task.ExpressionId)
	}

	_, err := db.ExecContext(ctx, q.String(), args...)
	if err != nil {
		return fmt.Errorf("failed to insert tasks: %w", err)
	}

	return nil
}

func GetExpressions(ctx context.Context, db *sql.DB) ([]*model.Expression, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, expression, status, result FROM expressions")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expressions: %w", err)
	}
	defer rows.Close()

	var expressions []*model.Expression
	for rows.Next() {
		var expr model.Expression
		if err := rows.Scan(&expr.ID, &expr.Expression, &expr.Status, &expr.Result); err != nil {
			return nil, fmt.Errorf("failed to scan expression: %w", err)
		}
		expressions = append(expressions, &expr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during rows iteration: %w", err)
	}

	return expressions, nil
}

func GetExpressionByID(ctx context.Context, db *sql.DB, id string) (*model.Expression, error) {
	var expr model.Expression
	query := `SELECT id, expression, status, result FROM expressions WHERE id = $1`
	err := db.QueryRowContext(ctx, query, id).Scan(
		&expr.ID,
		&expr.Expression,
		&expr.Status,
		&expr.Result,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expression not found")
		}
		return nil, fmt.Errorf("failed to get expression: %w", err)
	}
	return &expr, nil
}

func GetNextPendingTask(ctx context.Context, db *sql.DB) (*model.Task, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `SELECT id, arg1, arg2, operation, expression_id
			FROM tasks 
			WHERE result IS NULL 
			AND arg1 NOT GLOB '*[a-zA-Z]*' 
			AND arg2 NOT GLOB '*[a-zA-Z]*' 
			LIMIT 1;`

	var task model.Task
	err = tx.QueryRowContext(ctx, query).Scan(
		&task.ID,
		&task.Arg1,
		&task.Arg2,
		&task.Operation,
		&task.ExpressionId,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		fmt.Println(err)
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &task, nil
}

func UpdateTaskResult(ctx context.Context, db *sql.DB, taskID string, result float64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Обновляем результат задачи
	_, err = tx.ExecContext(ctx, `
        UPDATE tasks 
        SET result = $1, status = 'completed' 
        WHERE id = $2`,
		result, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task result: %w", err)
	}

	// 2. Получаем ID выражения
	var expressionID string
	err = tx.QueryRowContext(ctx, `
        SELECT expression_id FROM tasks WHERE id = $1`,
		taskID).Scan(&expressionID)
	if err != nil {
		return fmt.Errorf("failed to get expression ID: %w", err)
	}

	// 3. Обновляем зависимости в других задачах
	_, err = tx.ExecContext(ctx, `
        UPDATE tasks
        SET 
            arg1 = CASE 
                WHEN arg1 = $1 THEN $2::text 
                ELSE arg1 
            END,
            arg2 = CASE 
                WHEN arg2 = $1 THEN $2::text 
                ELSE arg2 
            END
        WHERE expression_id = $3 
        AND result IS NULL
        AND (arg1 = $1 OR arg2 = $1)`,
		taskID, fmt.Sprintf("%g", result), expressionID)
	if err != nil {
		return fmt.Errorf("failed to update dependent tasks: %w", err)
	}

	// 4. Проверяем, все ли задачи выражения выполнены
	var pendingTasks int
	err = tx.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM tasks 
        WHERE expression_id = $1 AND result IS NULL`,
		expressionID).Scan(&pendingTasks)
	if err != nil {
		return fmt.Errorf("failed to check pending tasks: %w", err)
	}

	// 5. Если все задачи выполнены - обновляем выражение
	if pendingTasks == 0 {
		var finalResult float64
		err := tx.QueryRowContext(ctx, `
			UPDATE expressions 
			SET status = 'completed', result = (
				SELECT tasks.result FROM tasks join expressions
				on tasks.expression_id = expressions.id 
				where expressions.result = tasks.id
				LIMIT 1
			) 
			WHERE id = $1
			RETURNING result`,
			expressionID).Scan(&finalResult)

		if err != nil {
			return fmt.Errorf("failed to update expression and get final result: %w", err)
		}
	}

	return tx.Commit()
}
