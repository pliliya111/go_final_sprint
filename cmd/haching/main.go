package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	var password string
	fmt.Print("Введите пароль для хэширования: ")
	fmt.Scanln(&password)

	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Fatalf("Ошибка хэширования пароля: %v", err)
	}

	fmt.Printf("Хэшированный пароль: %s\n", hashedPassword)

	var passwordToCheck string
	fmt.Print("Введите пароль для проверки: ")
	fmt.Scanln(&passwordToCheck)

	if checkPasswordHash(passwordToCheck, hashedPassword) {
		fmt.Println("Пароль верный!")
	} else {
		fmt.Println("Пароль неверный!")
	}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
