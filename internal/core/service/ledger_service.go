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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
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
						workerEvent	*event.WorkerEvent,
						) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		apiService: apiService,
		workerRepository: workerRepository,
		workerEvent: workerEvent,
	}
}

// About handle/convert http status code
func errorStatusCode(statusCode int, serviceName string) error{
	childLogger.Info().Str("func","errorStatusCode").Interface("serviceName", serviceName).Interface("statusCode", statusCode).Send()
	var err error
	switch statusCode {
		case http.StatusUnauthorized:
			err = erro.ErrUnauthorized
		case http.StatusForbidden:
			err = erro.ErrHTTPForbiden
		case http.StatusNotFound:
			err = erro.ErrNotFound
		default:
			err = errors.New(fmt.Sprintf("service %s in outage", serviceName))
		}
	return err
}

// About create a moviment transaction
func (s *WorkerService) MovimentTransaction(ctx context.Context, moviment model.Moviment) (*model.MovimentTransaction, error){
	childLogger.Info().Str("func","MovimentTransaction").Interface("trace-request-id", ctx.Value("trace-request-id")).Interface("moviment", moviment).Send()

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
			childLogger.Info().Interface("trace-request-id", trace_id ).Msg("ROLLBACK !!!!")
			err :=  s.workerEvent.WorkerKafka.AbortTransaction(ctx)
			if err != nil {
				childLogger.Info().Interface("trace-request-id", trace_id ).Err(err).Msg("failed to kafka AbortTransaction")
			}
			tx.Rollback(ctx)
		} else {
			childLogger.Info().Interface("trace-request-id", trace_id ).Msg("COMMIT !!!!")
			err =  s.workerEvent.WorkerKafka.CommitTransaction(ctx)
			if err != nil {
				childLogger.Info().Interface("trace-request-id", trace_id ).Err(err).Msg("Failed to Kafka CommitTransaction")
			}
			tx.Commit(ctx)
		}
		s.workerRepository.DatabasePGServer.ReleaseTx(conn)
		span.End()
	}()

	//set time transaction
	transaction_at := time.Now()

	// check transaction type
	transactionType := model.TransactionType{TransactionTypeID: moviment.Type}
	_, err = s.workerRepository.GetTransactionType(ctx, transactionType)
	if err != nil {
		return nil, err
	}

	// set the account from according transaction type (deposit / withdrawn)
	switch moviment.Type {
		case "DEPOSIT":
			moviment.AccountTo = &model.Account {AccountID: "ACC-BANK"}  // set account to 
		case "WITHDRAW":
			moviment.AccountTo = &model.Account {AccountID: "ACC-BANK"}  // set account to 
		case "WIRE_TRANSFER":
		default:
			return nil, erro.ErrTypeInvalid
	}
	
	// check account_to
	headers := map[string]string{
		"Content-Type":  	"application/json;charset=UTF-8",
		"X-Request-Id": 	trace_id,
		"x-apigw-api-id": 	s.apiService[0].XApigwApiId,
		"Host": 			s.apiService[0].HostName,
	}
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/get/" + moviment.AccountFrom.AccountID,
		Method: s.apiService[0].Method,
		Timeout: 15,
		Headers: &headers,
	}
	_, statusCode, err := apiService.CallRestApi(ctx,
												httpClient, 
												nil)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[0].Name)
	}

	// check account_from
	httpClient = go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/get/" + moviment.AccountFrom.AccountID,
		Method: s.apiService[0].Method,
		Timeout: 15,
		Headers: &headers,
	}

	_, statusCode, err = apiService.CallRestApi(ctx,
												httpClient, 
												nil)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[0].Name)
	}

	// create a ledger transaction
	transaction := model.Transaction{	Currency: moviment.Currency,
										Description: moviment.Type,
										TransactionAt: transaction_at,
									}
	res_transaction, err := s.workerRepository.AddTransaction(ctx, tx, transaction)
	if err != nil {
		return nil, err
	}

	// add transaction detail - double partition
	var debitAmount, creditAmount int64
	var account_id string
	ledger_id := "BANK:LIABILITY"

	list_transactionDetail := []model.TransactionDetail{}
	for i := 0; i < 2; i++ {
		switch moviment.Type {
			case "DEPOSIT":
				if i == 0 {
					creditAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					debitAmount = 0
					account_id = moviment.AccountFrom.AccountID
				} else {
					creditAmount = 0
					debitAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					account_id = moviment.AccountTo.AccountID
				}
			case "WITHDRAW":
				if i == 0 {
					creditAmount = 0
					debitAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					account_id = moviment.AccountFrom.AccountID
					} else {
					creditAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					debitAmount = 0
					account_id = moviment.AccountTo.AccountID
				}
			case "WIRE_TRANSFER":
				if i == 0 {
					creditAmount = 0
					debitAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					account_id = moviment.AccountFrom.AccountID	
				} else {
					creditAmount = int64(moviment.Amount * 100) // store as an integer (2 decimal position)
					debitAmount = 0
					account_id = moviment.AccountTo.AccountID
				}
			default:
				return nil, erro.ErrTypeInvalid
		}
		transactionDetail := model.TransactionDetail{	FkTxID: 		res_transaction.ID,
														TransactionAt: 	transaction_at,
														AccountID:		account_id,
														FkLedgerID: 	ledger_id,
														DebitAmount: 	debitAmount,
														CreditAmount: 	creditAmount,
														TraceID: 		trace_id,}

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
			childLogger.Error().Interface("trace-request-id", trace_id ).Err(err).Msg("failed to kafka BeginTransaction")
			return nil, err
		}
		
		// Prepare to event
		key := strconv.Itoa(res_transaction.ID)
		payload_bytes, err := json.Marshal(res_moviment_transaction)
		if err != nil {
			return nil, err
		}
		
		// prepare header
		carrier := propagation.MapCarrier{}
		otel.GetTextMapPropagator().Inject(ctx, &carrier)
	
		headers_msk := make(map[string]string)
		for k, v := range carrier {
			headers_msk[k] = v
		}

		spanContext := span.SpanContext()
		headers["trace-request-id"] = trace_id
		headers["TraceID"] = spanContext.TraceID().String()
		headers["SpanID"] = spanContext.SpanID().String()

		// publish event
		err = s.workerEvent.WorkerKafka.Producer(s.workerEvent.Topics[0], key, &headers, payload_bytes)
		if err != nil {
			return nil, err
		}
		res_moviment_transaction.Status = "event sended via kafka"
	} else {
		res_moviment_transaction.Status = "event not send, kafka unabled"
	}
												
	return &res_moviment_transaction, nil
}

// About get ledger all moviment
func (s *WorkerService) GetAccountStament(ctx context.Context, moviment model.Moviment) (*model.MovimentStatement, error){
	childLogger.Info().Str("func","GetAccountStament").Interface("trace-request-id", ctx.Value("trace-request-id")).Interface("moviment", moviment).Send()

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

// About get ledger moviment per data
func (s *WorkerService) GetAccountMovimentStatementPerDate(ctx context.Context, moviment model.Moviment) (*model.MovimentStatement, error){
	childLogger.Info().Str("func","GetAccountMovimentStatementPerDate").Interface("trace-request-id", ctx.Value("trace-request-id")).Interface("moviment", moviment).Send()

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