package service

import(
	"fmt"
	"time"
	"context"
	"errors"
	"net/http"
	"strconv"
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/go-ledger/internal/core/model"
	"github.com/go-ledger/internal/core/erro"
	"github.com/go-ledger/internal/adapter/database"
	"github.com/go-ledger/internal/adapter/event"
	
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("component","go-ledger").Str("package","internal.core.service").Logger()
var apiService go_core_api.ApiService

type WorkerService struct {
	apiService			[]model.ApiService
	workerRepository 	*database.WorkerRepository
	workerEvent			*event.WorkerEvent
}

// About create a new worker service
func NewWorkerService(	workerRepository *database.WorkerRepository,
						apiService		[]model.ApiService,
						workerEvent	*event.WorkerEvent,) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		apiService: apiService,
		workerRepository: workerRepository,
		workerEvent: workerEvent,
	}
}

// About handle/convert http status code
func errorStatusCode(statusCode int) error{
	var err error
	switch statusCode {
		case http.StatusUnauthorized:
			err = erro.ErrUnauthorized
		case http.StatusForbidden:
			err = erro.ErrHTTPForbiden
		case http.StatusNotFound:
			err = erro.ErrNotFound
		default:
			err = erro.ErrServer
		}
	return err
}

// About create a moviment transaction
func (s *WorkerService) MovimentTransaction(ctx context.Context, moviment model.Moviment) (*model.MovimentTransaction, error){
	childLogger.Info().Str("func","MovimentTransaction").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("moviment", moviment).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.MovimentTransaction")
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))
	
	// prepare database
	tx, conn, err := s.workerRepository.DatabasePGServer.StartTx(ctx)
	if err != nil {
		return nil, err
	}
	
	// handle connection
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		} else {
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	//business rule
	transaction_at := time.Now()
	
	// Get the Account ID (PK) from Account-service
	// Set headers
	headers := map[string]string{
		"Content-Type":  "application/json;charset=UTF-8",
		"X-Request-Id": trace_id,
		"x-apigw-api-id": s.apiService[0].XApigwApiId,
		"Host": s.apiService[0].HostName,
	}
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/get/" + moviment.AccountID,
		Method: s.apiService[0].Method,
		Timeout: 15,
		Headers: &headers,
	}

	res_payload, statusCode, err := apiService.CallRestApi(	ctx,
															httpClient, 
															nil)
	if err != nil {
		return nil, errorStatusCode(statusCode)
	}

	jsonString, err  := json.Marshal(res_payload)
	if err != nil {
		return nil, errors.New(err.Error())
    }
	var account_parsed model.Account
	json.Unmarshal(jsonString, &account_parsed)

	// get/chech transaction type
	transactionType := model.TransactionType{TransactionTypeID: moviment.Type}
	_, err = s.workerRepository.GetTransactionType(ctx, transactionType)
	if err != nil {
		return nil, err
	}

	// add transaction
	transaction := model.Transaction{	Currency: moviment.Currency,
										Description: moviment.Type,
										TransactionAt: transaction_at,
									}
	res_transaction, err := s.workerRepository.AddTransaction(ctx, tx, transaction)
	if err != nil {
		return nil, err
	}

	// add transaction detail - double partition
	var debitAmount int64  = 0
	var creditAmount int64  = 0
	account_id := moviment.AccountID
	ledger_id := "BANK:LIABILITY"

	list_transactionDetail := []model.TransactionDetail{}
	for i := 0; i < 2; i++ {

		switch moviment.Type {
			case "DEPOSIT":
				if i == 0 {
					creditAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					debitAmount = 0
				} else {
					creditAmount = 0
					debitAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					account_id = "ACC-BANK"
				}
			case "WITHDRAW":
				if i == 0 {
					creditAmount = 0
					debitAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
				} else {
					creditAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					debitAmount = 0
					account_id = "ACC-BANK"
				}
			default:
				return nil, erro.ErrTypeInvalid
		}
		transactionDetail := model.TransactionDetail{	FkTxID: 		res_transaction.ID,
														TransactionAt: 	transaction_at,
														AccountID:		account_id,
														FkLedgerID: 	ledger_id,
														DebitAmount: 	debitAmount,
														CreditAmount: 	creditAmount,}

		res_transactionDetail, err := s.workerRepository.AddTransactionDetail(ctx, tx, transactionDetail)
		if err != nil {
			return nil, err
		}

		list_transactionDetail = append(list_transactionDetail, *res_transactionDetail)
	}

	res_moviment_transaction := model.MovimentTransaction{	Transaction: *res_transaction,
															TransactionDetail: list_transactionDetail, }

	// --------------------------- STEP 02 ------------------------------------//
	// Start Kafka transaction (if workerEvent is nil means kafka unabled)
	if s.workerEvent != nil {
		err = s.workerEvent.WorkerKafka.BeginTransaction()
		if err != nil {
			childLogger.Error().Interface("trace-resquest-id", trace_id ).Err(err).Msg("failed to kafka BeginTransaction")
			return nil, err
		}
		// Prepare to event
		key := strconv.Itoa(res_transaction.ID)
		payload_bytes, err := json.Marshal(res_moviment_transaction)
		if err != nil {
			return nil, err
		}
		// publish event
		err = s.workerEvent.WorkerKafka.Producer(ctx, s.workerEvent.Topics[0], key, &trace_id, payload_bytes)
		if err != nil {
			return nil, err
		}
		res_moviment_transaction.Status = "event sended via kafka"
	} else {
		res_moviment_transaction.Status = "event not send, kafka unabled"
	}
												
	return &res_moviment_transaction, nil
}

// About get account statement
func (s *WorkerService) GetAccountStament(ctx context.Context, moviment model.Moviment) (*model.MovimentStatement, error){
	childLogger.Info().Str("func","GetAccountStament").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("moviment", moviment).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.GetAccountStament")
	defer span.End()

	// get the summary of all moviments
	moviment_summary, err := s.workerRepository.GetSumMovimentAccount(ctx, moviment)
	if err != nil {
		return nil, err
	}
	// adjust the decimal position
	for i, _ := range *moviment_summary{
		(*moviment_summary)[i].Amount = (*moviment_summary)[i].Amount / 100
	}

	// get the details of moviments
	moviment_statement, err := s.workerRepository.GetAllMovimentAccount(ctx, moviment)
	if err != nil {
		return nil, err
	}
	// adjust the decimal position
	for i, _ := range *moviment_statement{
		(*moviment_statement)[i].Amount = (*moviment_statement)[i].Amount / 100
	}
	
	res_moviment_statement := model.MovimentStatement{
		MovimentSummary: moviment_summary,
		MovimentStatement: moviment_statement,
	}
	
	return &res_moviment_statement, nil
}

// About get account statement
func (s *WorkerService) GetAccountMovimentStatementPerDate(ctx context.Context, moviment model.Moviment) (*model.MovimentStatement, error){
	childLogger.Info().Str("func","GetAccountMovimentStatementPerDate").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("moviment", moviment).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.GetAccountMovimentStatementPerDate")
	defer span.End()

	// get the summary of all moviments
	moviment_summary, err := s.workerRepository.GetSumMovimentAccountPerDate(ctx, moviment)
	if err != nil {
		return nil, err
	}
	// adjust the decimal position
	for i, _ := range *moviment_summary{
		(*moviment_summary)[i].Amount = (*moviment_summary)[i].Amount / 100
	}

	// get the details of moviments
	moviment_statement, err := s.workerRepository.GetAllMovimentAccountPerDate(ctx, moviment)
	if err != nil {
		return nil, err
	}
	// adjust the decimal position
	for i, _ := range *moviment_statement{
		(*moviment_statement)[i].Amount = (*moviment_statement)[i].Amount / 100
	}
	
	res_moviment_statement := model.MovimentStatement{
		MovimentSummary: moviment_summary,
		MovimentStatement: moviment_statement,
	}
	
	return &res_moviment_statement, nil
}