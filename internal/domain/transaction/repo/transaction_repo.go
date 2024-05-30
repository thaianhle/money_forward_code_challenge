package repo

import (
	"context"
	"money_forward_code_challenge/internal/domain/transaction/aggregate"
	"money_forward_code_challenge/internal/domain/transaction/models"
)

type Query struct {
	SortBy string `form:"sort_by"`
	Order  string `form:"order_by"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type TransactionRepo[TxTypeT any] interface {
	Create(context.Context, *models.Transaction, TxTypeT) error
	Update(context.Context, *models.Transaction, TxTypeT) error
	Delete(context.Context, *models.Transaction, TxTypeT) error
	GetByUserId(context.Context, uint32, *Query) ([]*aggregate.TransactionByDetails, error)
	GetByAccountId(context.Context, uint32, *Query) ([]*aggregate.TransactionByDetails, error)
	GetById(context.Context, uint32) (*aggregate.TransactionByDetails, error)
	BeginTx() TxTypeT
}

type TransactionCacheRepo interface {
	Set(context.Context, *aggregate.TransactionByDetails) error
	Delete(context.Context, uint32) error
	GetById(context.Context, uint32) (*aggregate.TransactionByDetails, error)
	GetByAccountId(context.Context, uint32, *Query) ([]*aggregate.TransactionByDetails, error)
}
