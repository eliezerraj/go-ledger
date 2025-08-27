package service

import(
	"fmt"
	"time"
	"context"
	"errors"
	"net/http"
	"strconv"
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/go-ledger/internal/core/model"
	"github.com/go-ledger/internal/core/erro"
	"github.com/go-ledger/internal/adapter/database"
	"github.com/go-ledger/internal/adapter/event"

	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_api "github.com/eliezerraj/go-core/api"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("component","go-ledger").Str("package","internal.core.service").Logger()
var apiService 	go_core_api.ApiService

type WorkerService struct {
	goCoreRestApiService	go_core_api.ApiService
	apiService				[]model.ApiService
	workerRepository 		*database.WorkerRepository
	workerEvent				*event.WorkerEvent
	mutex    				sync.Mutex
}

// About create a new worker service
func NewWorkerService(	goCoreRestApiService	go_core_api.ApiService,	
						workerRepository 		*database.WorkerRepository,
						apiService				[]model.ApiService,
						workerEvent				*event.WorkerEvent,) *WorkerService{
	childLogger.Info().Str("func","NewWorkerService").Send()

	return &WorkerService{
		goCoreRestApiService: goCoreRestApiService,
		apiService: apiService,
		workerRepository: workerRepository,
		workerEvent: workerEvent,
		mutex:    sync.Mutex{},
	}
}

// About handle/convert http status code
func errorStatusCode(statusCode int, serviceName string, msg_err error) error{
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
			err = errors.New(fmt.Sprintf("service %s in outage => cause error: %s", serviceName, msg_err.Error() ))
		}
	return err
}

// About handle/convert http status code
func (s *WorkerService) Stat(ctx context.Context) (go_core_pg.PoolStats){
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	return s.workerRepository.Stat(ctx)
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
	defer s.workerRepository.DatabasePGServer.ReleaseTx(conn)

	// handle connection
	defer func() {
		if err != nil {
			childLogger.Info().Interface("trace-request-id", trace_id ).Msg("ROLLBACK TX !!!")
			tx.Rollback(ctx)
		} else {
			childLogger.Info().Interface("trace-request-id", trace_id ).Msg("COMMIT TX !!!")
			tx.Commit(ctx)
		}
		span.End()
	}()

	//set time transaction
	transaction_at := time.Now()

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
	
	// check transaction type
	transactionType := model.TransactionType{TransactionTypeID: moviment.Type}
	_, err = s.workerRepository.GetTransactionType(ctx, transactionType)
	if err != nil {
		return nil, err
	}

	// check account_from
	headers := map[string]string{
		"Content-Type":  	"application/json;charset=UTF-8",
		"X-Request-Id": 	trace_id,
		"x-apigw-api-id": 	s.apiService[0].XApigwApiId,
		"Host": 			s.apiService[0].HostName,
	}
	httpClient := go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/get/" + moviment.AccountFrom.AccountID,
		Method: s.apiService[0].Method,
		Timeout: s.apiService[0].HttpTimeout,
		Headers: &headers,
	}
	_, statusCode, err := apiService.CallRestApiV1(	ctx,
													s.goCoreRestApiService.Client,
													httpClient, 
													nil)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[0].Name, err)
	}

	// check account_from
	httpClient = go_core_api.HttpClient {
		Url: 	s.apiService[0].Url + "/get/" + moviment.AccountTo.AccountID,
		Method: s.apiService[0].Method,
		Timeout: s.apiService[0].HttpTimeout,
		Headers: &headers,
	}

	_, statusCode, err = apiService.CallRestApiV1(	ctx,
													s.goCoreRestApiService.Client,
													httpClient, 
													nil)
	if err != nil {
		return nil, errorStatusCode(statusCode, s.apiService[0].Name, err)
	}

	// create a ledger transaction
	transaction := model.Transaction{	Currency: moviment.Currency,
										Description: moviment.Type,
										TransactionAt: transaction_at,}
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
		err = s.ProducerEventKafka(ctx, moviment, res_moviment_transaction, *res_transaction)
		if err != nil {
			return nil, err
		}
		res_moviment_transaction.Status = "event sended via kafka"
	} else {
		res_moviment_transaction.Status = "event not send, kafka unabled"
	}
	
	return &res_moviment_transaction, nil
}

// About producer a event in kafka
func(s *WorkerService) ProducerEventKafka(ctx context.Context, moviment model.Moviment, res_moviment_transaction model.MovimentTransaction, res_transaction model.Transaction) (err error) {
	childLogger.Info().Str("func","ProducerEventKafka").Interface("trace-request-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "service.ProducerEventKafka")
	defer span.End()
	
	trace_id := fmt.Sprintf("%v",ctx.Value("trace-request-id"))

	// create a mutex to avoid commit over a transaction on air
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a transacrion
	err = s.workerEvent.WorkerKafka.BeginTransaction()
	if err != nil {
		childLogger.Error().Interface("trace-request-id", trace_id ).Err(err).Msg("failed to kafka BeginTransaction")
		// Create a new producer and start a transaction
		err = s.workerEvent.DestroyWorkerEventProducerTx(ctx)
		if err != nil {
			return  err
		}
		s.workerEvent.WorkerKafka.BeginTransaction()
		if err != nil {
			return err
		}
		childLogger.Info().Interface("trace-request-id", trace_id ).Msg("success to recreate a new producer")
	}

	// Prepare to event
	key := strconv.Itoa(res_transaction.ID)
	payload_bytes, err := json.Marshal(res_moviment_transaction)
	if err != nil {
		return err
	}
		
	// prepare header
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, &carrier)
	
	headers_msk := make(map[string]string)
	for k, v := range carrier {
		headers_msk[k] = v
	}

	spanContext := span.SpanContext()
	headers_msk["trace-request-id"] = trace_id
	headers_msk["TraceID"] = spanContext.TraceID().String()
	headers_msk["SpanID"] = spanContext.SpanID().String()

	// publish event
	err = s.workerEvent.WorkerKafka.Producer(s.workerEvent.Topics[0], key, &headers_msk, payload_bytes)
		
	//force a error SIMULARTION
	if(trace_id == "force-rollback"){
		err = erro.ErrForceRollback
	}
	
	if err != nil {
		childLogger.Err(err).Interface("trace-request-id", trace_id ).Msg("KAFKA ROLLBACK !!!")
		err_msk := s.workerEvent.WorkerKafka.AbortTransaction(ctx)
		if err_msk != nil {
			childLogger.Err(err_msk).Interface("trace-request-id", trace_id ).Msg("failed to kafka AbortTransaction")
			return err_msk
		}
		return err
	}

	err = s.workerEvent.WorkerKafka.CommitTransaction(ctx)
	if err != nil {
		childLogger.Err(err).Interface("trace-request-id", trace_id ).Msg("Failed to Kafka CommitTransaction = KAFKA ROLLBACK COMMIT !!!")
		err_msk := s.workerEvent.WorkerKafka.AbortTransaction(ctx)
		if err_msk != nil {
			childLogger.Err(err_msk).Interface("trace-request-id", trace_id ).Msg("failed to kafka AbortTransaction during CommitTransaction")
			return err_msk
		}
		return err
	}

	childLogger.Info().Interface("trace-request-id", trace_id ).Msg("KAFKA COMMIT !!!")

    return 
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