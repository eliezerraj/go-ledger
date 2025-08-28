package api

import (
	"fmt"
	"time"
	"context"
	"encoding/json"
	"reflect"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/go-ledger/internal/core/service"
	"github.com/go-ledger/internal/core/model"
	"github.com/go-ledger/internal/core/erro"

	"github.com/eliezerraj/go-core/coreJson"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_tools "github.com/eliezerraj/go-core/tools"
)

var (
	childLogger = log.With().Str("component", "go-ledger").Str("package", "internal.adapter.api").Logger()
	core_json 		coreJson.CoreJson
	core_apiError 	coreJson.APIError
	core_tools 		go_core_tools.ToolsCore
	tracerProvider 	go_core_observ.TracerProvider
)

type HttpRouters struct {
	workerService 	*service.WorkerService
	ctxTimeout		time.Duration
}

// Above create routers
func NewHttpRouters(workerService *service.WorkerService,
					ctxTimeout	time.Duration) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
		ctxTimeout: ctxTimeout,
	}
}

// About return a health
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Health").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Live").Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Header").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About show all context values
func (h *HttpRouters) Context(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Context").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	contextValues := reflect.ValueOf(req.Context()).Elem()
	json.NewEncoder(rw).Encode(fmt.Sprintf("%v",contextValues))
}

// About show pgx stats
func (h *HttpRouters) Stat(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Stat").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	res := h.workerService.Stat(req.Context())

	json.NewEncoder(rw).Encode(res)
}

// About handle error
func (h *HttpRouters) ErrorHandler(trace_id string, err error) *coreJson.APIError {
	if strings.Contains(err.Error(), "context deadline exceeded") {
    	err = erro.ErrTimeout
	} 
	switch err {
	case erro.ErrBadRequest:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusBadRequest)
	case erro.ErrNotFound:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusNotFound)
	case erro.ErrTimeout:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusGatewayTimeout)
	default:
		core_apiError = core_apiError.NewAPIError(err, trace_id, http.StatusInternalServerError)
	}
	return &core_apiError
}

// About add a moviment into ledger
func (h *HttpRouters) MovimentTransaction(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","MovimentTransaction").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	//context
	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	//trace
	span := tracerProvider.Span(ctx, "adapter.api.MovimentTransaction")
	defer span.End()

	trace_id := fmt.Sprintf("%v", ctx.Value("trace-request-id"))

	moviment := model.Moviment{}
	err := json.NewDecoder(req.Body).Decode(&moviment)
    if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
    }
	defer req.Body.Close()

	res, err := h.workerService.MovimentTransaction(ctx, moviment)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get account statement
func (h *HttpRouters) GetAccountMovimentStatement(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccountMovimentStatement").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	//context
	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	// trace
	span := tracerProvider.Span(ctx, "adapter.api.GetAccountMovimentStatement")
	defer span.End()

	trace_id := fmt.Sprintf("%v", ctx.Value("trace-request-id"))

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	moviment := model.Moviment{}
	moviment.AccountFrom = model.Account{AccountID: varID}

	// call service
	res, err := h.workerService.GetAccountStament(ctx, moviment)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get account statement
func (h *HttpRouters) GetAccountMovimentStatementPerDate(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccountMovimentStatementPerDate").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	//context
	ctx, cancel := context.WithTimeout(req.Context(), h.ctxTimeout * time.Second)
    defer cancel()

	// trace
	span := tracerProvider.Span(ctx, "adapter.api.GetAccountMovimentStatementPerDate")
	defer span.End()

	trace_id := fmt.Sprintf("%v", ctx.Value("trace-request-id"))

	//parameters
	params := req.URL.Query()
	varAcc := params.Get("account-id")
	varDate := params.Get("date_start")

	convertDate, err := core_tools.ConvertToDate(varDate)
	if err != nil {
		return h.ErrorHandler(trace_id, erro.ErrBadRequest)
	}

	moviment := model.Moviment{ AccountFrom: model.Account{AccountID: varAcc},
								TransactionAt: convertDate}

	// call service
	res, err := h.workerService.GetAccountMovimentStatementPerDate(ctx, moviment)
	if err != nil {
		return h.ErrorHandler(trace_id, err)
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}