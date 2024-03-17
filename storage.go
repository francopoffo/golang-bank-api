package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(id int, account *UpdateAccountRequest) error
	GetAccounts() ([]*Account, error)
	GetAccountById(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=admin sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{db: db}, nil
}

// Init initializes the PostgresStore.
func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

// createAccountTable creates the accounts table if it does not exist.
func (s *PostgresStore) createAccountTable() error {
	query := `CREATE TABLE IF NOT EXISTS accounts (
		id SERIAL PRIMARY KEY,
		first_name VARCHAR(50) NOT NULL,
		last_name VARCHAR(50) NOT NULL,
		number BIGINT NOT NULL UNIQUE,
		balance BIGINT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	)`

	_, err := s.db.Exec(query)

	return err
}

func (s *PostgresStore) CreateAccount(account *Account) error {
	query := `INSERT INTO accounts (first_name, last_name, number, balance, created_at, updated_at) 
	VALUES ($1, $2, $3, $4)`

	resp, err := s.db.Exec(
		query,
		account.FirstName,
		account.LastName,
		account.Number,
		account.Balance,
		account.CreatedAt,
		account.UpdatedAt)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)

	return nil
}

func (s *PostgresStore) DeleteAccount(id int) error {
	query := "DELETE FROM accounts WHERE id = $1"

	resp, err := s.db.Exec(query, id)

	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", resp)

	return nil
}

func (s *PostgresStore) UpdateAccount(id int, account *UpdateAccountRequest) error {
	// Create a buffer to dynamically build the query
	var queryBuffer bytes.Buffer
	queryBuffer.WriteString("UPDATE accounts SET ")

	// Keep track if any field is updated
	var updatedFields []interface{}

	// Keep track if any field is provided in the request
	var hasFieldsToUpdate bool

	// Check if first name is provided
	if account.FirstName != "" {
		queryBuffer.WriteString("first_name = $1, ")
		updatedFields = append(updatedFields, account.FirstName)
		hasFieldsToUpdate = true
	}

	// Check if last name is provided
	if account.LastName != "" {
		queryBuffer.WriteString("last_name = $2, ")
		updatedFields = append(updatedFields, account.LastName)
		hasFieldsToUpdate = true
	}

	// Remove trailing comma and space
	if hasFieldsToUpdate {
		query := queryBuffer.String()[:queryBuffer.Len()-2] // Remove the last comma and space
		query += "updated_at = NOW() WHERE id = $3"         // Append the WHERE clause

		// Execute the dynamic query
		_, err := s.db.Exec(query, append(updatedFields, id)...)
		if err != nil {
			return err
		}
		return nil
	}

	// If no fields are provided in the request
	return errors.New("no fields provided for update")
}

func (s *PostgresStore) GetAccountById(id int) (*Account, error) {
	rows, err := s.db.Query("SELECT * FROM accounts WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}
	return nil, fmt.Errorf("account with id %d not found", id)
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := s.db.Query("SELECT * FROM accounts")
	if err != nil {
		return nil, err
	}
	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := &Account{}
	err := rows.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt)

	return account, err
}
