package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provides all functions to execute database queries individually or in transactions.
type Store struct {
	*Queries
	db DBTX
}

// NewStore creates a new Store with the given DBTX.
func NewStore(db DBTX) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(*Queries) error) error {
	// Assert store.db to a type that supports Begin, like pgxpool.Pool
	conn, ok := store.db.(*pgxpool.Pool)
	if !ok {
		return fmt.Errorf("store.db is not a pgxpool.Pool")
	}

	// Begin a transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a transaction-specific Queries instance
	qtx := store.Queries.WithTx(tx)

	// Execute the provided function
	err = fn(qtx)
	if err != nil {
		// Rollback on error
		rbErr := tx.Rollback(ctx)
		if rbErr != nil {
			return fmt.Errorf("transaction error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	// Commit the transaction
	if commitErr := tx.Commit(ctx); commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

type TransferTxParams struct {
	FromAccountID int   `json:"from_account_id"`
	ToAccountID   int   `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   int      `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Create a transfer record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: int64(arg.FromAccountID),
			ToAccountID:   int64(arg.ToAccountID),
			Amount:        int64(arg.Amount),
		})
		if err != nil {
			return err
		}

		// Create a debit entry for the sender
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: int64(arg.FromAccountID),
			Amount:    -int64(arg.Amount), // Negative for debit
		})
		if err != nil {
			return err
		}

		// Create a credit entry for the recipient
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: int64(arg.ToAccountID),
			Amount:    int64(arg.Amount), // Positive for credit
		})
		if err != nil {
			return err
		}

		// TODO: update code

		return nil
	})

	return result, err
}
