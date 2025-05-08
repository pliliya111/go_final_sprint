package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pliliya111/go_final_sprint/internal/model"
	"golang.org/x/crypto/bcrypt"
)

func isUserExists(ctx context.Context, db *sql.DB, username string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE name = $1", username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}
	return count > 0, nil
}
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
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

	// hashedPassword, err := bcrypt.GenerateFromPassword(
	// 	[]byte(user.Password),
	// 	bcrypt.DefaultCost,
	// )
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to hash password: %w", err)
	// }
	// hashedPassword, err := hashPassword(user.Password)
	// if err != nil {
	// 	return 0, fmt.Errorf("failed to hash password: %w", err)
	// }

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

	// Сравниваем пароль
	// if err := bcrypt.CompareHashAndPassword(
	// 	[]byte(password),
	// 	[]byte(user.Password),
	// ); err != nil {
	// 	fmt.Println(err)
	// 	return nil, errors.New("invalid password")
	// }
	// hashedPassword, err := hashPassword(password)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to hash password: %w", err)
	// }
	// if checkPasswordHash(hashedPassword, user.Password) {
	// 	fmt.Println("Пароль верный!")
	// } else {
	// 	fmt.Println("Пароль неверный!")
	// 	fmt.Printf("password %s не такой как в бд %s", hashedPassword, user.Password)
	// 	return nil, fmt.Errorf("password %s не такой как в бд %s", hashedPassword, user.Password)
	// }
	if user.Password != password {
		return nil, fmt.Errorf("")
	}
	// Возвращаем пользователя БЕЗ пароля
	return &model.User{
		ID:   user.ID,
		Name: user.Name,
	}, nil
}
