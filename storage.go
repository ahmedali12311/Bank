package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type storage interface {
	createAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountbyID(int) (*Account, error)
	GetAllAccounts() (*[]Account, error)
	GetAccountbyNumber(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostGresStore() (*PostgresStore, error) {
	connStr := "user=ellie dbname=bankaccount password=091093Aa sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.CreateAccountTable()

}

func (s *PostgresStore) CreateAccountTable() error {
	query := `Create table if not exists BankAccount(
	id serial primary key ,
	first_name varchar(30),
	last_name varchar(50),
	number int,
	encrypted_password varchar(250),
	balance int,
	created_at timestamp
)`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStore) createAccount(acc *Account) error {
	Query := `insert into bankaccount(first_name,last_name,number,encrypted_password,balance,created_at)values($1,$2,$3,$4,$5,$6)`

	resp, err := s.db.Query(Query, acc.First_Name, acc.Last_Name, acc.Number, acc.EncryptedPassowrd, acc.Balance, acc.Created_at)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)
	return nil
}

func (s *PostgresStore) UpdateAccount(acc *Account) error {
	Query := `update bankaccount set balance=$1 where id=$2`
	_, err := s.db.Query(Query, acc.Balance, acc.ID)

	return err
}

func (s *PostgresStore) DeleteAccount(id int) error {
	Query := `delete from bankaccount where id=$1`
	_, err := s.db.Query(Query, id)

	return err

}
func (s *PostgresStore) GetAccountbyNumber(number int) (*Account, error) {
	rows, err := s.db.Query(`Select * from bankaccount where number=$1`, number)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("Account with number [%d] not found", number)
}
func (s *PostgresStore) GetAccountbyID(id int) (*Account, error) {
	rows, err := s.db.Query(`Select * from bankaccount where id=$1`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("Account %d not found", id)
}

func (s *PostgresStore) GetAllAccounts() (*[]Account, error) {
	rows, err := s.db.Query("SELECT * FROM bankaccount")
	if err != nil {
		log.Fatal(err)
	}
	var account []Account
	for rows.Next() {

		acc, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		account = append(account, *acc)

	}
	return &account, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	acc := new(Account)
	err := rows.Scan(&acc.ID, &acc.First_Name, &acc.Last_Name, &acc.Number, &acc.EncryptedPassowrd, &acc.Balance, &acc.Created_at)
	return acc, err

}
