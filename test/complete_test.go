package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"money_forward_code_challenge/internal/common/httpresponse"
	"money_forward_code_challenge/internal/domain/transaction/usecase/transaction"
	"net/http"
	"testing"
)

func doCreateTransaction(t *testing.T, reqData *transaction.CreateReq, userId uint32) *httpresponse.Response {

	baseUrl := "http://localhost:8080/api/users/" + fmt.Sprint(userId) + "/transactions"
	reqBytes, err := json.Marshal(reqData)
	if err != nil {
		fmt.Println(err)
	}

	res, err := http.Post(baseUrl, "application/json", bytes.NewBuffer(reqBytes))

	if err != nil {
		t.Errorf("error sending request: %v", err)
	}

	defer res.Body.Close()
	var response *httpresponse.Response
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Errorf("error decoding response: %v", err)
		return nil
	}

	return response
}

func doGetTransaction(t *testing.T, userId uint32, queryParams map[string]interface{}) *httpresponse.Response {
	baseUrl := "http://localhost:8080/api/users/" + fmt.Sprint(userId) + "/transactions"

	l := 0
	for param, value := range queryParams {
		if l == 0 {
			baseUrl += "?"
		}
		if l > 0 {
			baseUrl += "&"
		}
		baseUrl += fmt.Sprintf("%s=%v", param, value)
		l += 1
	}

	res, err := http.Get(baseUrl)
	if err != nil {
		t.Errorf("error sending request: %v", err)
	}

	defer res.Body.Close()
	var response *httpresponse.Response
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		t.Errorf("error decoding response: %v", err)
		return nil
	}

	return response
}

func TestGetTransactions(t *testing.T) {
	{
		// not found user
		var userId uint32 = 100
		res := doGetTransaction(t, userId, nil)
		t.Log(res)
		if res.Code == http.StatusNotFound {
			t.Log("1. Passed", res.ErrCodeString)
		}
	}
	{
		// not found account on query transactions
		var userId uint32 = 1 // in initdb
		res := doGetTransaction(t, userId, map[string]interface{}{
			"account_id": 123, // support
			"limit":      5,   // support
			"offset":     0,   // support
		})
		if res.Code == http.StatusNotFound {
			t.Log("2. Passed", res.ErrCodeString)
		}
	}
	{
		// success but not have any transactions
		var userId uint32 = 1 // in initdb
		res := doGetTransaction(t, userId, map[string]interface{}{
			"account_id": 1, // support
			"limit":      5, // support
			"offset":     0, // support
		})
		if res.Code == http.StatusOK {
			t.Log("3. Passed", res.ErrCodeString)
		}
	}
	{
		// create some transaction before get transactions
		var userId uint32 = 1
		amountRands := []float32{15000, 23000, 26000, 27000, 28000, 29500, 32420}
		transactionRands := []string{"deposit", "withdraw"}
		nums_tran := 10
		for i := 0; i < nums_tran; i++ {
			amount := amountRands[rand.Intn(len(amountRands))]
			tranTyp := transactionRands[rand.Intn(len(transactionRands))]
			doCreateTransaction(t, &transaction.CreateReq{
				AccountId:       3,
				Amount:          amount,
				TransactionType: tranTyp,
			}, userId)
		}

		res := doGetTransaction(t, userId, map[string]interface{}{
			"account_id": 3,
			"limit":      10,
			"offset":     0,
		})

		data, err := json.Marshal(res.Data)
		if err != nil {
			t.Log("4. Error Parsed Json", err.Error())
		}

		if res.Code == http.StatusOK {
			t.Log("4. Passed", string(data))
		}
	}
}

func TestCreateTransactions(t *testing.T) {
	{
		// not valid because transaction_type not must be one of [deposit, withdraw]
		reqData := transaction.CreateReq{
			AccountId:       123,
			Amount:          100.5,
			TransactionType: "deposit-1",
		}

		res := doCreateTransaction(t, &reqData, 1)
		if res.Code == http.StatusBadRequest {
			t.Log("1. Passed", res.ErrCodeString)
		}
	}

	{
		// not valid because Amount in [Min, Max]
		reqData := transaction.CreateReq{
			AccountId:       123,
			Amount:          100.5,
			TransactionType: "deposit",
		}

		res := doCreateTransaction(t, &reqData, 1)
		if res.Code == http.StatusBadRequest {
			t.Log("2. Passed", res.ErrCodeString)
		}
	}

	{
		// not valid because account_id = 10
		// not exist in database 404
		reqData := transaction.CreateReq{
			AccountId:       10,
			Amount:          15200,
			TransactionType: "withdraw",
		}

		res := doCreateTransaction(t, &reqData, 1)
		if res.Code == http.StatusNotFound {
			t.Log("3. Passed", res.ErrCodeString)
		}
	}
	{
		// not valid because account_id = 1
		// exist but owner user_id not must be [2] as in url param
		reqData := transaction.CreateReq{
			AccountId:       1,
			Amount:          15000,
			TransactionType: "withdraw",
		}

		res := doCreateTransaction(t, &reqData, 2) // user_id is 2 (not valid)
		if res.Code == http.StatusBadRequest {
			t.Log("4. Passed", res.ErrCodeString)
		}
	}

	{
		// user_id I init db
		// but account_id = 1 only have balance = 1000
		// so failed on withdraw 15000
		req := transaction.CreateReq{
			AccountId:       1,
			Amount:          15000,
			TransactionType: "withdraw",
		}
		res := doCreateTransaction(t, &req, 1) // same user_id as url_param

		// but balance account 1 = 1000 which isn't enough for withdraw amount = 15000
		if res.Code == http.StatusBadRequest {
			t.Log("5. Passed", res.ErrCodeString)
		}
	}

	{

		// user_id I init db
		// but account_id = 3 have balance = 300000
		// so success on withdraw 50000
		req := transaction.CreateReq{
			AccountId:       3,
			Amount:          50000,
			TransactionType: "withdraw",
		}
		res := doCreateTransaction(t, &req, 1) // same user_id as url_param
		if res.Code == http.StatusCreated {
			t.Log("6. Passed: ", res.Data)
		}
	}
}
