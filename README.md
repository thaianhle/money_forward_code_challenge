## PROJECT MF_CODE_CHALLENGE
### Directory Structure
```go
├── cmd
│   ├── configuration  // as glue layer to emsemble composite-usecase, repo
│   │   │   ├── monolithic // concrete glue layer (monolithic)
│   │   │   ├── microservice // concrete glue layer (microservice)
├── internal // source code for project
│   ├── common // folder shared
│   │   ├── composite // folder used from configuration layer, it compose all usecase interface and repo interface
│   │   │   ├── transaction_composite.go // as name (get all usecase and composite into each class to use with each service layer for configuration)
│   │   ├── exception // folder implement for some exception to construct and check invalid (transaction_type, amount_range_min_max ...)
│   │   │   ├── invalid_err_transaction_type.go
│   │   │   ├── ...
│   │   ├── httpresponse // folder used for responses, I implemented transform http code on each error or success (include data message)
// transform with explicit error or success data 
func (r *Response) TransformToCreatedSuccess(dataMessage any) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusCreated
	return r.constructDataMessage(dataMessage)
}

func (r *Response) TransformToUpdatedSuccess(dataMessage any) *Response {
	r.resetBeforeTransform()
	r.Code = http.StatusAccepted
	return r.constructDataMessage(dataMessage)
}
│   ├── domain
│   │   ├── transaction // transaction as domain bounded context
│   │   │   ├── aggregate // contain models aggregate across models folder
│   │   │   │   ├──transaction_details.go // include transaction and account fields if be needed
│   │   │   │   ├──account_details.go // include account and user fields if be needed
│   │   │   ├── models
│   │   │   │   ├──account.go // root account model
│   │   │   │   ├──transaction.go // root transaction model
│   │   │   │   ├──user.go // root user model
│   │   │   └── repo
│   │   │   │   ├── transaction_repo.go // interfaces (one for PersistentRepo, one for CacheRepo)
│   │   │   │   ├── user_repo.go // interfaces (one for PersistentRepo, one for CacheRepo)
│   │   │   └── usecase // core layer use repo interfaces
│   │   │   │   ├── transaction // all usecase interfaces for each action used
│   │   │   │   │   ├── create_transaction.go // include (CreateReq, CreateUseCase, defaultCreateUseCaseImpl, ConstructorNew)
│   │   │   │   │   ├── ...

How `TransactionService` use one usecase for example: 
`CreateUseCase` interface. It performs the following steps:

1. Creates a new `Transaction` model from the `CreateReq` data.
2. Persists the transaction to the persistent repository using the provided transaction type.
3. Creates a `TransactionByDetails` object with the transaction details.
4. Formats the date in the `TransactionByDetails` object.
5. Asynchronously updates the cache repository with the `TransactionByDetails` object using the worker pool.
6. Returns the `TransactionByDetails` object, a `Job` object representing the asynchronous update, and any errors that occurred.

below is code in `create_transaction.go`

type CreateReq struct {
	// userId, BankTypeName not required in json binding
	// because in url api
	// it be checked
	UserId          uint32  `json:"user_id"`
	BankType        string  `json:"bank_type"`
	AccountId       uint32  `json:"account_id,binding:required"`
	Amount          float32 `json:"amount,binding:required"`
	TransactionType string  `json:"transaction_type,binding:required"`
}

type CreateUseCase[TxType any] interface {
	Execute(ctx context.Context, req *CreateReq, tx TxType) (
		*aggregate.TransactionByDetails,
		*repo_pool_async.Job,
		error)
}

type defaultCreateUseCase[TxType any] struct {
	persistentRepo repo.TransactionRepo[TxType]
	cacheRepo      repo.TransactionCacheRepo
	pool           *repo_pool_async.RepoUpdatePoolBusyWaiting
	logger         *zap.Logger
}

func NewCreateUseCase[TxType any](persistentRepo repo.TransactionRepo[TxType],
	cacheRepo repo.TransactionCacheRepo, logger *zap.Logger,
	poolSizeWorker int,
) CreateUseCase[TxType] {
	return &defaultCreateUseCase[TxType]{
		logger:         logger,
		persistentRepo: persistentRepo,
		cacheRepo:      cacheRepo,
		pool:           repo_pool_async.NewPool(context.TODO(), poolSizeWorker, logger),
		//pool:           repo_pool_async.NewRepoUpdatePoolBusyWaiting[repo.TransactionCacheRepo, *aggregate.TransactionByDetails](),
	}
}

// implement usecase interface
func (d *defaultCreateUseCase[TxType]) Execute(ctx context.Context, req *CreateReq, tx TxType) (
	*aggregate.TransactionByDetails,
	*repo_pool_async.Job,
	error)
	

│   │   │   │   │   ├── ...
│   │   │   │   ├── user // all usecase interfaces for each action used same as transaction
│   │   │   │   │   ├── ...
│   │   │   │   │   ├── ...
│   └── pkgs
│       └── repo_pool_async // this is library I implemented used for async update after acid transaction
// repo_pool_async -- it mean when we
// make transaction type = deposit
// then update balance account with account_id = 3
// now if use redis cache to update within usecase layer
// seem challenge to rollback redis cache if one in two transaction within persistent repo be failed
// I use technique return asyncJob on each usecase have related transaction
// then from outside core layer
// we can use asyncJob to run set cache in redis
// if have any error asyncJob will never be updated
// because worker gouroutine pool only process asyncJob
// on condition reponseTime <= expiredTime
this example in transaction_service.go 
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
```


The code follows a typical use case pattern, separating the transaction creation logic from the persistence and caching operations. It also utilizes asynchronous updates to improve performance by offloading the cache update to a worker pool.
## Usage
1. Run executable script ./deploy.sh
```bash
chmod +x ./deploy.sh
```

### 2. Run Docker Compose
- docker compose file: `docker-compose.yaml` 
- deploy transaction service: `deploy/monolithic/Dockerfile.transaction.service`
- launch new terminal in root folder project, and run:
```bash
./deploy.sh run_dev
```
### 3. Test CreateTransaction and GetTransaction
- TestCreateTransaction (6 case passed)
- TestGetTransactions (4 case passed)
- launch new terminal in root folder, and run:
```bash
cd test
```
```bash
go test -v -run TestCreateTransactions
go test -v -run TestGetTransactions
```

### 4. Free to POSTMAN:
#### API Endpoints
#### a. Create Transaction
**URL:** `/api/v1/<user_id>/transactions`

**Method:** `POST`
**URL Parameters:**
- `user_id`: The ID of the user for whom the transaction is being created.
**Request Body:**

```json
{
  "account_id": 456,
  "amount": 100.50,
  "transaction_type": "deposit"
}
```

**Response Data:**
```json
{
   "code": 201 | 400 | 404 ...
   "err_code_string: // anything for detail error
   "data": {
   // empty if have any error
   // return transaction details if success
   }
}
```
#### b. Get Transactions

**URL:** `/api/v1/:user_id/transactions`

**Method:** `GET`

**URL Parameters:**

- `user_id`: The ID of the user whose transactions are being retrieved.

**Query Parameters:**

- `account_id`: The ID of the account to filter transactions by. If not provided, all transactions for the user will be returned.
- `limit`: The Limit support for paginate
- `offset`: The Offset support for paginate

**Description:**

This endpoint retrieves transactions based on the provided parameters:

- If no `account_id` query parameter is provided, it will return all transactions for the specified `user_id`.
- If an `account_id` query parameter is provided, it will return transactions for the specified `user_id` and `account_id`.

**Response Data:**

```json
{
"code": 200 ok | 400 bad request | 404 not found account or not user_id owner ...
   "err_code_string: // anything for detail error
   "data": {
   // empty if have any error
   // return transaction details if success
   // example success data
       [
         {
           "id": 1,
           "user_id": 123,
           "account_id": 456,
           "amount": 100.50,
           "transaction_type": "deposit"|"withdraw",
           "bank": "ACB"|"VIB"|"VCB"
           "created_at": "2023-05-30 12:34:56 +0700 UTC"
         },
         {
           "id": 2,
           "user_id": 123,
           "account_id": 789,
           "amount": 50.25,
           "transaction_type": "deposit"|"withdraw",
           "bank": "ACB"|"VIB"|"VCB",
           "created_at": "2023-06-01 09:15:30 +0700 UTC"
         }
       ]
   }
}
```

#### c. Delete Transaction
### Delete Transaction

**URL:** `/api/v1/:user_id/transactions

**Method:** `DELETE`

**Request Body**
```json
{
   "account_id": 456,
   "transaction_id": 10,
}
```
**Response Data:**

```json
{
   "code": accepted 202 | bad request 400 | 404 not found account or not user_id owner or not found transaction_id...
   "err_code_string: // anything for detail error
   "data": {
   // empty if have any error
   // return transaction details if success
   }
}
```