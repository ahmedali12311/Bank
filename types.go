package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type loginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}
type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

type TransferRequest struct {
	ToAccount int `json:"toAccount"`
	Amount    int `json:"amount"`
}
type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"LastName"`
	Password  string `json:"password"`
}

type Account struct {
	ID                int
	First_Name        string
	Last_Name         string
	Number            int
	EncryptedPassowrd string `json:"-"`

	Balance    int
	Created_at time.Time
}

func (a *Account) ValidPassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassowrd), []byte(pw)) == nil
}
func NewAccount(FirstName, LastName, password string) (*Account, error) {
	encpw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:                rand.Intn(10000),
		First_Name:        FirstName,
		Last_Name:         LastName,
		Number:            rand.Intn(10000),
		EncryptedPassowrd: string(encpw),
		Balance:           rand.Intn(10000),
		Created_at:        time.Now(),
	}, nil
}
