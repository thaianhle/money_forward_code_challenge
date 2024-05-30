package composite

import (
	"gorm.io/gorm"
	"money_forward_code_challenge/internal/domain/transaction/repo"
	transaction_usecase "money_forward_code_challenge/internal/domain/transaction/usecase/transaction"
	user_usecase "money_forward_code_challenge/internal/domain/transaction/usecase/user"
)

type TransactionRepoComposite struct {
	PersistentRepo repo.TransactionRepo[*gorm.DB]
	CacheRepo      repo.TransactionCacheRepo
}

type UserRepoComposite struct {
	PersistentRepo repo.UserRepo[*gorm.DB]
	CacheRepo      repo.UserCacheRepo
}

type UserUseCaseComposite struct {
	GetAccountByAccountId user_usecase.GetAccountByAccountId[*gorm.DB]
	UpdateBalanceAccount  user_usecase.UpdateBalanceAccountUseCase[*gorm.DB]
	GetUserById           user_usecase.GetUserById[*gorm.DB]
}

type TransactionUseCaseComposite struct {
	Create             transaction_usecase.CreateUseCase[*gorm.DB]
	Delete             transaction_usecase.DeleteTransactionById[*gorm.DB]
	GetByAccountId     transaction_usecase.GetTransactionsByAccountId[*gorm.DB]
	GetByTransactionId transaction_usecase.GetTransactionById[*gorm.DB]
	GetByUserId        transaction_usecase.GetTransactionsByUserId[*gorm.DB]
}
