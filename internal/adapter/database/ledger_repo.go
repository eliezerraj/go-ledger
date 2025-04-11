package database

import (
	"context"
	"time"
	"errors"
	
	"github.com/go-ledger/internal/core/model"
	"github.com/go-ledger/internal/core/erro"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

var tracerProvider go_core_observ.TracerProvider
var childLogger = log.With().Str("component","go-ledger").Str("package","internal.adapter.database").Logger()

type WorkerRepository struct {
	DatabasePGServer *go_core_pg.DatabasePGServer
}

// Above new worker
func NewWorkerRepository(databasePGServer *go_core_pg.DatabasePGServer) *WorkerRepository{
	childLogger.Info().Str("func","NewWorkerRepository").Send()

	return &WorkerRepository{
		DatabasePGServer: databasePGServer,
	}
}

// Above add transaction
func (w WorkerRepository) AddTransaction(ctx context.Context, tx pgx.Tx, transaction model.Transaction) (*model.Transaction, error){
	childLogger.Info().Str("func","AddTransaction").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "database.AddTransaction")
	defer span.End()

	// prepare
	transaction.TransactionAt = time.Now()

	//query
	query := `INSERT INTO transaction (	currency,
										description,
 										transaction_at) 
										VALUES($1, $2, $3) RETURNING id`
	
	// execute	
	row := tx.QueryRow(ctx, query,	transaction.Currency,
									transaction.Description,
									transaction.TransactionAt,
									)

	var id int
	
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	transaction.ID = id
	
	return &transaction, nil
}

// Above add transactionDetail
func (w WorkerRepository) AddTransactionDetail(ctx context.Context, tx pgx.Tx, transactionDetail model.TransactionDetail) (*model.TransactionDetail, error){
	childLogger.Info().Str("func","AddTransactionDetail").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Interface("transactionDetail",transactionDetail).Send()

	// trace
	ctx, span := tracerProvider.SpanCtx(ctx, "database.AddTransactionDetail")
	defer span.End()

	// prepare
	transactionDetail.CreatedAt = time.Now()

	//query
	query := `INSERT INTO transaction_detail (	fk_transaction_id,
												fk_ledger_id, 
												fk_account_id,
												debit_amount,
												credit_amount,
												created_at) 
												VALUES($1, $2, $3, $4, $5, $6) RETURNING id`
	
	// execute	
	row := tx.QueryRow(ctx, query,  transactionDetail.FkTxID,  
									transactionDetail.FkLedgerID,
									transactionDetail.AccountID,
									transactionDetail.DebitAmount,
									transactionDetail.CreditAmount,
									transactionDetail.CreatedAt)

	var id int
	
	if err := row.Scan(&id); err != nil {
		return nil, errors.New(err.Error())
	}

	transactionDetail.ID = id
	
	return &transactionDetail, nil
}

// Above add transaction
func (w WorkerRepository) GetTransactionType(ctx context.Context, transactionType model.TransactionType) (*model.TransactionType, error){
	childLogger.Info().Str("func","GetTransactionType").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetTransactionType")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	res_transaction_type := model.TransactionType{}

	query := `SELECT transaction_id,
					 description
				FROM public.transaction_type 
				WHERE transaction_id = $1`

	// execute			
	rows, err := conn.Query(ctx, query, transactionType.TransactionTypeID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&res_transaction_type.TransactionTypeID,
							&res_transaction_type.Description,)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		return &res_transaction_type, nil
	}
	
	return nil, erro.ErrNotFound
}

// Above get the sum of all moviment (summary)
func (w WorkerRepository) GetSumMovimentAccount(ctx context.Context, moviment model.Moviment) (*[]model.Moviment, error){
	childLogger.Info().Str("func","GetSumMovimentAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetSumMovimentAccount")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	list_moviment := []model.Moviment{}
	query := `select tr.currency,
					 sum((td.credit_amount - td.debit_amount)) as amount
				from transaction tr,
					transaction_detail td
				where tr.id  = td.fk_transaction_id
				and td.fk_account_id = $1
				group by tr.currency`

	// execute			
	rows, err := conn.Query(ctx, query, moviment.AccountFrom.AccountID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&moviment.Currency,
							&moviment.Amount)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		list_moviment = append(list_moviment, moviment)
	}
	
	return &list_moviment, nil
}

// Above get all mobviments done in an account
func (w WorkerRepository) GetAllMovimentAccount(ctx context.Context, moviment model.Moviment) (*[]model.Moviment, error){
	childLogger.Info().Str("func","GetAllMovimentAccount").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetAllMovimentAccount")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	list_moviment := []model.Moviment{}
	// prepare query
	query := `select tr.currency,
					tr.transaction_at,
					tr.description,
					(td.credit_amount - td.debit_amount) as Amount
				from transaction tr,
					transaction_detail td
				where tr.id  = td.fk_transaction_id
				and td.fk_account_id = $1
				order by td.fk_transaction_id desc, td.id desc`

	// execute			
	rows, err := conn.Query(ctx, query, moviment.AccountFrom.AccountID)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&moviment.Currency,
							&moviment.TransactionAt,
							&moviment.Type,
							&moviment.Amount)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		list_moviment = append(list_moviment, moviment)
	}
	return &list_moviment, nil
}

// Above get the sum of all moviment (summary)
func (w WorkerRepository) GetSumMovimentAccountPerDate(ctx context.Context, moviment model.Moviment) (*[]model.Moviment, error){
	childLogger.Info().Str("func","GetSumMovimentAccountPerDate").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetSumMovimentAccountPerDate")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	// prepare query
	list_moviment := []model.Moviment{}
	query := `select tr.currency,
					 sum((td.credit_amount - td.debit_amount)) as amount
				from transaction tr,
					 transaction_detail td
				where tr.id  = td.fk_transaction_id
				and td.fk_account_id = $1
				and tr.transaction_at >= $2
				group by tr.currency`

	// execute			
	rows, err := conn.Query(ctx, query, moviment.AccountFrom.AccountID, moviment.TransactionAt)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&moviment.Currency,
							&moviment.Amount)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		list_moviment = append(list_moviment, moviment)
	}
	
	return &list_moviment, nil
}

// Above get all mobviments done in an account
func (w WorkerRepository) GetAllMovimentAccountPerDate(ctx context.Context, moviment model.Moviment) (*[]model.Moviment, error){
	childLogger.Info().Str("func","GetAllMovimentAccountPerDate").Interface("trace-resquest-id", ctx.Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(ctx, "database.GetAllMovimentAccountPerDate")
	defer span.End()

	// prepare database
	conn, err := w.DatabasePGServer.Acquire(ctx)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer w.DatabasePGServer.Release(conn)

	list_moviment := []model.Moviment{}
	// prepare query
	query := `select tr.currency,
					 tr.transaction_at,
					 tr.description,
					 (td.credit_amount - td.debit_amount) as Amount
				from transaction tr,
				  	 transaction_detail td
				where tr.id  = td.fk_transaction_id
				and td.fk_account_id = $1
				and tr.transaction_at >= $2
				order by td.fk_transaction_id desc, td.id desc`

	// execute			
	rows, err := conn.Query(ctx, query, moviment.AccountFrom.AccountID, moviment.TransactionAt)
	if err != nil {
		return nil, errors.New(err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan( 	&moviment.Currency,
							&moviment.TransactionAt,
							&moviment.Type,
							&moviment.Amount)
		if err != nil {
			return nil, errors.New(err.Error())
        }
		list_moviment = append(list_moviment, moviment)
	}
	return &list_moviment, nil
}