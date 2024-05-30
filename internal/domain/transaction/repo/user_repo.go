package repo

import (
	"context"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/models"
)

type UserRepo[TxType any] interface {
	CreateUser(ctx context.Context, user_model *models.User, tx TxType) error
	CreateAccount(ctx context.Context, account_model *models.Account, tx TxType) error
	UpdateBalance(ctx context.Context, account_id uint32, new_balance float32, tx TxType) error
	UpdateAccount(ctx context.Context, account_model *models.Account, tx TxType) error
	DeleteAccountById(ctx context.Context, account_id uint32, tx TxType) error
	GetUserById(ctx context.Context, user_id uint32) (*models.User, error)
	GetAccountByAccountId(ctx context.Context, account_id uint32) (*aggregate.AccountByDetails, error)
	BeginTx() TxType
}

type UserCacheRepo interface {
	CreateUser(context.Context, *models.User) error
	SetAccount(context.Context, *aggregate.AccountByDetails) error
	DeleteAccountById(context.Context, uint32) error
	GetUserById(ctx context.Context, user_id uint32) (*models.User, error)
	GetAccountByAccountId(ctx context.Context, account_id uint32) (*aggregate.AccountByDetails, error)
}
