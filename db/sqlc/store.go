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

type txKey string

const transactionContextKey txKey = "transaction_name"

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
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// txName := ctx.Value(transactionContextKey)

		// fmt.Println(txName, "Create transfer")
		// Create a transfer record
		result.Transfer, err = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: int64(arg.FromAccountID),
			ToAccountID:   int64(arg.ToAccountID),
			Amount:        int64(arg.Amount),
		})
		if err != nil {
			return err
		}

		// fmt.Println(txName, "Create entry 1")
		// Create a debit entry for the sender
		result.FromEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: int64(arg.FromAccountID),
			Amount:    -int64(arg.Amount), // Negative for debit
		})
		if err != nil {
			return err
		}

		// fmt.Println(txName, "Create entry 2")
		// Create a credit entry for the recipient
		result.ToEntry, err = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: int64(arg.ToAccountID),
			Amount:    int64(arg.Amount), // Positive for credit
		})
		if err != nil {
			return err
		}

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, _ = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			result.ToAccount, result.FromAccount, _ = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)

		}
		return nil
	})

	return result, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1 int64,
	amount1 int64,
	accountID2 int64,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID:     accountID2,
		Amount: amount2,
	})

	return

}
