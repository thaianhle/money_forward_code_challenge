package monolithic

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"money_forward_code_challenge/internal/common/composite"
	exception "money_forward_code_challenge/internal/common/exception"
	"money_forward_code_challenge/internal/common/httpresponse"
	"money_forward_code_challenge/internal/domain/transaction/models"
	"money_forward_code_challenge/internal/domain/transaction/usecase/transaction"
	transactionusecase "money_forward_code_challenge/internal/domain/transaction/usecase/transaction"
	userusecase "money_forward_code_challenge/internal/domain/transaction/usecase/user"
)

type TransactionService struct {
	repo struct {
		transaction *composite.TransactionRepoComposite
		user        *composite.UserRepoComposite
	}
	useCase struct {
		transaction *composite.TransactionUseCaseComposite
		user        *composite.UserUseCaseComposite
	}
	logger *zap.Logger
}

func NewTransactionService(transactionRepoComposite *composite.TransactionRepoComposite, userRepoComposite *composite.UserRepoComposite, logger *zap.Logger, poolSizeWorkerUseCase int) *TransactionService {
	return &TransactionService{
		logger: logger,
		repo: struct {
			transaction *composite.TransactionRepoComposite
			user        *composite.UserRepoComposite
		}{
			transaction: transactionRepoComposite,
			user:        userRepoComposite,
		},
		useCase: struct {
			transaction *composite.TransactionUseCaseComposite
			user        *composite.UserUseCaseComposite
		}{
			transaction: &composite.TransactionUseCaseComposite{
				Create:             transactionusecase.NewCreateUseCase(transactionRepoComposite.PersistentRepo, transactionRepoComposite.CacheRepo, logger, poolSizeWorkerUseCase),
				Delete:             transactionusecase.NewDeleteTransactionByAccountId(transactionRepoComposite.PersistentRepo, transactionRepoComposite.CacheRepo, logger, poolSizeWorkerUseCase),
				GetByTransactionId: transactionusecase.NewDefaultGetTransactionById(transactionRepoComposite.PersistentRepo, transactionRepoComposite.CacheRepo, logger),
				GetByUserId:        transactionusecase.NewGetTransactionsByUserId(transactionRepoComposite.PersistentRepo, transactionRepoComposite.CacheRepo, logger),
				GetByAccountId:     transactionusecase.NewDefaultGetTransactionsByAccountId(transactionRepoComposite.PersistentRepo, transactionRepoComposite.CacheRepo, logger),
			},
			user: &composite.UserUseCaseComposite{
				GetUserById:           userusecase.NewGetUserByIdUseCase(userRepoComposite.PersistentRepo, userRepoComposite.CacheRepo, logger),
				GetAccountByAccountId: userusecase.NewGetAccountByAccountId(userRepoComposite.PersistentRepo, userRepoComposite.CacheRepo, logger),
				UpdateBalanceAccount:  userusecase.NewUpdateBalanceAccountUseCase(userRepoComposite.PersistentRepo, userRepoComposite.CacheRepo, logger, poolSizeWorkerUseCase),
			},
		},
	}
}

func (t *TransactionService) createTransactionByUser(ctx context.Context, req *transaction.CreateReq) *httpresponse.Response {
	res := &httpresponse.Response{}
	userId := getUserIdFromContext(ctx)

	transactionTypeErrCheck := exception.NewCheckExceptionTransactionType([]string{"deposit", "withdraw"})

	transactionTypeErrCheck.Check(string(req.TransactionType))
	if transactionTypeErrCheck.Error() != "" {
		return res.TransformToBadRequest(transactionTypeErrCheck.Error())
	}

	transactionMinMaxAmountErrCheck := exception.NewCheckErrAmountValue(10000, 20000000)
	transactionMinMaxAmountErrCheck.Check(req.Amount)
	if transactionMinMaxAmountErrCheck.Error() != "" {
		return res.TransformToBadRequest(transactionMinMaxAmountErrCheck.Error())
	}

	// get account check balance
	accountDetail, err := t.useCase.user.GetAccountByAccountId.Execute(ctx, &userusecase.GetAccountByAccountIdReq{
		AccountId: req.AccountId,
	})

	if err != nil {
		// account not found or user account owner is not same as url param <user_id>
		return res.TransformToNotFound(err.Error())
	}

	if accountDetail.UserId != userId {
		// user account owner is not same as url param <user_id>
		return res.TransformToBadRequest("user account owner is not same as url param <user_id>")
	}

	if req.TransactionType == models.TRANSACTIONTYPEWITHDRAW {
		if accountDetail.Balance < req.Amount {
			return res.TransformToBadRequest("balance is not enough")
		}
	}

	// pass into user_id,bank_type to update detail transaction if save success
	// then redis cache will have all details include account fields
	// and user_id
	// because I want create transaction
	// but I don't want must join to get details
	// on create
	req.UserId = userId
	req.BankType = accountDetail.Bank
	// open session tx pointer, to control from outside
	sessionTx := t.repo.transaction.PersistentRepo.BeginTx()

	// create transaction and return detail model
	transactionDetail, asyncJobCreateTransaction, err := t.useCase.transaction.Create.Execute(ctx, req, sessionTx)
	if err != nil {
		sessionTx.Rollback()
		return res.TransformToInternalServerError(err.Error())
	}

	// update balance
	asyncJobUpdateBalance, err := t.useCase.user.UpdateBalanceAccount.Execute(ctx, &userusecase.UpdateBalanceAccountReq{
		AccountId:       req.AccountId,
		OldBalance:      accountDetail.Balance,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
	}, sessionTx)

	if err != nil {
		_ = sessionTx.Rollback().Error
		return res.TransformToInternalServerError(err.Error())
	}

	err = sessionTx.Commit().Error

	if err != nil {
		_ = sessionTx.Rollback().Error
		return res.TransformToInternalServerError(err.Error())
	}

	defer func(ctx context.Context) {
		// run all update async for create transaction and update balance
		// run method
		// get response time from async job
		// it not called then, it simplify don't process from own pool
		// this is because can panic before reach code here
		asyncJobCreateTransaction.Run(ctx)
		asyncJobUpdateBalance.Run(ctx)
	}(ctx)

	return res.TransformToCreatedSuccess(transactionDetail)
}

func (t *TransactionService) getTransactionsByUserId(ctx context.Context, req *transactionusecase.GetTransactionByUserIdReq) *httpresponse.Response {
	res := &httpresponse.Response{}
	_, err := t.useCase.user.GetUserById.Execute(ctx, &userusecase.GetUserByIdReq{UserId: req.UserId})
	if err != nil {
		return res.TransformToNotFound(err.Error())
	}

	transactionModels, err := t.useCase.transaction.GetByUserId.Execute(ctx, req)
	if err != nil {
		return res.TransformToInternalServerError(err.Error())
	}

	return res.TransformToSuccessOk(transactionModels)
}

func (t *TransactionService) getTransactionsByAccountId(ctx context.Context, req *transactionusecase.GetTransactionByAccountIdReq) *httpresponse.Response {
	res := &httpresponse.Response{}
	userId := getUserIdFromContext(ctx)
	accountDetail, err := t.useCase.user.GetAccountByAccountId.Execute(ctx, &userusecase.GetAccountByAccountIdReq{
		AccountId: req.AccountId,
	})
	if err != nil {
		// not found account before get transactions
		// or found account but user owner is not right
		return res.TransformToNotFound(err.Error())
	}

	if accountDetail.UserId != userId {
		// user account owner is not same as url param <user_id>
		return res.TransformToBadRequest("user account owner is not same as url param <user_id>")
	}
	transactionModels, err := t.useCase.transaction.GetByAccountId.Execute(ctx, req)
	if err != nil {
		return res.TransformToInternalServerError(err.Error())
	}

	return res.TransformToSuccessOk(transactionModels)
}

func (t *TransactionService) deleteTransactionByUser(ctx *gin.Context, req *transactionusecase.DeleteReq) interface{} {
	res := &httpresponse.Response{}
	userId := getUserIdFromContext(ctx)

	accountDetail, err := t.useCase.user.GetAccountByAccountId.Execute(ctx, &userusecase.GetAccountByAccountIdReq{
		AccountId: req.AccountId,
	})

	if err != nil {
		// account not found or user account owner is not same as url param <user_id>
		return res.TransformToNotFound(err.Error())
	}

	if accountDetail.UserId != userId {
		// user account owner is not same as url param <user_id>
		return res.TransformToBadRequest("user account owner is not same as url param <user_id>")
	}

	// open session tx pointer, to control from outside
	sessionTx := t.repo.transaction.PersistentRepo.BeginTx()

	// create transaction and return detail model
	transactionDetail, asyncJobDeleteTransaction, err := t.useCase.transaction.Delete.Execute(ctx, req, sessionTx)
	if err != nil {
		sessionTx.Rollback()
		return res.TransformToInternalServerError(err.Error())
	}

	// update balance
	asyncJobUpdateBalance, err := t.useCase.user.UpdateBalanceAccount.Execute(ctx, &userusecase.UpdateBalanceAccountReq{
		AccountId:       req.AccountId,
		OldBalance:      accountDetail.Balance,
		Amount:          transactionDetail.Amount,
		TransactionType: transactionDetail.TransactionType,
	}, sessionTx)

	if err != nil {
		sessionTx.Rollback()
		return res.TransformToInternalServerError(err.Error())
	}

	sessionTx.Commit()

	defer func(ctx context.Context) {
		asyncJobDeleteTransaction.Run(ctx)
		asyncJobUpdateBalance.Run(ctx)
	}(ctx)

	// delete should return status code
	// try to recommend me

	// status accepted
	return res.TransformToDeletedSuccess(transactionDetail)
}
