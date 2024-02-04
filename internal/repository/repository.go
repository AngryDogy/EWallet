package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"infotecs/internal/models"
)

type Repository interface {
	Connect() error
	Close() error
	Initialize() error
	BeginTransaction() error
	RollbackTransaction()
	CommitTransaction()
	SaveWallet(wallet *models.Wallet) (string, error)
	SaveTransaction(transaction *models.Transaction) (string, error)
	GetWallet(id string) (*models.Wallet, error)
	GetTransactionHistory(id string) (*[]models.Transaction, error)
	UpdateWallet(wallet *models.Wallet) error
}
type PostgresRepository struct {
	source string
	conn   *sql.DB
	tx     *sql.Tx
}

func NewPostgresRepository(username, pass, host, port, dbname string) Repository {
	pqdb := &PostgresRepository{source: fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", username, pass, host, port, dbname)}
	return pqdb
}
func (r *PostgresRepository) Connect() error {
	conn, err := sql.Open("postgres", r.source)
	if err != nil {
		return err
	}

	if err := conn.Ping(); err != nil {
		return err
	}
	r.conn = conn

	return nil
}

func (r *PostgresRepository) Close() error {
	return r.conn.Close()
}

func (r *PostgresRepository) Initialize() error {
	_, err := r.conn.Exec("CREATE TABLE IF NOT EXISTS wallets(id varchar PRIMARY KEY, balance numeric NOT NULL )")
	if err != nil {
		return err
	}

	_, err = r.conn.Exec("CREATE TABLE IF NOT EXISTS transactions(id varchar PRIMARY KEY, time timestamp NOT NULL , from_wallet varchar NOT NULL , to_wallet varchar NOT NULL , amount numeric NOT NULL )")
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) BeginTransaction() error {
	var err error
	r.tx, err = r.conn.Begin()
	return err

}

func (r *PostgresRepository) RollbackTransaction() {
	r.tx.Rollback()
	r.tx = nil
}

func (r *PostgresRepository) CommitTransaction() {
	r.tx.Commit()
}

func (r *PostgresRepository) SaveWallet(wallet *models.Wallet) (string, error) {
	if r.tx == nil {
		return "", errors.New("can't save an entity without an opened transaction")
	}

	wallet.Id = uuid.New().String()
	query := "INSERT INTO wallets(id, balance) VALUES ($1, $2) RETURNING id"
	var id uuid.UUID
	err := r.tx.QueryRow(query, wallet.Id, wallet.Balance).Scan(&id)
	if err != nil {
		return "", err
	}

	return id.String(), nil
}

func (r *PostgresRepository) SaveTransaction(transaction *models.Transaction) (string, error) {
	if r.tx == nil {
		return "", errors.New("can't save an entity without an opened transaction")
	}

	transaction.Id = uuid.New().String()
	query := "INSERT INTO transactions(id, time, from_wallet, to_wallet, amount) VALUES ($1, $2, $3, $4, $5) RETURNING id"
	var id string
	err := r.conn.QueryRow(query, transaction.Id, transaction.Time.Time, transaction.From, transaction.To, transaction.Amount).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *PostgresRepository) GetWallet(id string) (*models.Wallet, error) {
	query := "SELECT id, balance FROM wallets WHERE id = $1"
	var wallet models.Wallet
	err := r.conn.QueryRow(query, id).Scan(&wallet.Id, &wallet.Balance)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (r *PostgresRepository) GetTransactionHistory(id string) (*[]models.Transaction, error) {
	query := "SELECT id, time, from_wallet, to_wallet, amount FROM transactions WHERE from_wallet = $1 OR to_wallet = $1 ORDER BY time"
	var transactions []models.Transaction
	rows, err := r.conn.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		transaction.Time = &models.CustomTime{}
		err := rows.Scan(&transaction.Id, &transaction.Time.Time, &transaction.From, &transaction.To, &transaction.Amount)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return &transactions, nil
}

func (r *PostgresRepository) UpdateWallet(wallet *models.Wallet) error {
	if r.tx == nil {
		return errors.New("can't update an entity without an opened transaction")
	}

	query := "UPDATE wallets SET balance = $1 WHERE id = $2"
	rows, err := r.tx.Query(query, wallet.Balance, wallet.Id)
	defer rows.Close()

	return err
}
