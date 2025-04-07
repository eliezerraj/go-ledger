package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/rs/zerolog/log"

	"github.com/go-ledger/internal/core/service"
	"github.com/go-ledger/internal/core/model"
	"github.com/go-ledger/internal/core/erro"

	"github.com/eliezerraj/go-core/coreJson"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_tools "github.com/eliezerraj/go-core/tools"
)

var childLogger = log.With().Str("component", "go-ledger").Str("package", "internal.adapter.api").Logger()

var core_json 		coreJson.CoreJson
var core_apiError 	coreJson.APIError
var core_tools 		go_core_tools.ToolsCore
var tracerProvider 	go_core_observ.TracerProvider

type HttpRouters struct {
	workerService 	*service.WorkerService
}

// Above create routers
func NewHttpRouters(workerService *service.WorkerService) HttpRouters {
	childLogger.Info().Str("func","NewHttpRouters").Send()

	return HttpRouters{
		workerService: workerService,
	}
}

// About return a health
func (h *HttpRouters) Health(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Msg("Health")

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About return a live
func (h *HttpRouters) Live(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Live").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	json.NewEncoder(rw).Encode(model.MessageRouter{Message: "true"})
}

// About show all header received
func (h *HttpRouters) Header(rw http.ResponseWriter, req *http.Request) {
	childLogger.Info().Str("func","Header").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	json.NewEncoder(rw).Encode(req.Header)
}

// About add a moviment into ledger
func (h *HttpRouters) MovimentTransaction(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","MovimentTransaction").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()
	
	span := tracerProvider.Span(req.Context(), "adapter.api.MovimentTransaction")
	defer span.End()

	moviment := model.Moviment{}
	err := json.NewDecoder(req.Body).Decode(&moviment)
    if err != nil {
		core_apiError = core_apiError.NewAPIError(err, http.StatusBadRequest)
		return &core_apiError
    }
	defer req.Body.Close()

	res, err := h.workerService.MovimentTransaction(req.Context(), moviment)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get account statement
func (h *HttpRouters) GetAccountMovimentStatement(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccountMovimentStatement").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.GetAccountMovimentStatement")
	defer span.End()

	//parameters
	vars := mux.Vars(req)
	varID := vars["id"]

	moviment := model.Moviment{}
	moviment.AccountID = varID

	// call service
	res, err := h.workerService.GetAccountStament(req.Context(), moviment)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}

// About get account statement
func (h *HttpRouters) GetAccountMovimentStatementPerDate(rw http.ResponseWriter, req *http.Request) error {
	childLogger.Info().Str("func","GetAccountMovimentStatementPerDate").Interface("trace-resquest-id", req.Context().Value("trace-request-id")).Send()

	// trace
	span := tracerProvider.Span(req.Context(), "adapter.api.GetAccountMovimentStatementPerDate")
	defer span.End()

	//parameters
	params := req.URL.Query()
	varAcc := params.Get("account-id")
	varDate := params.Get("date_start")

	convertDate, err := core_tools.ConvertToDate(varDate)
	if err != nil {
		core_apiError = core_apiError.NewAPIError(erro.ErrUnmarshal, http.StatusBadRequest)
		return &core_apiError
	}

	moviment := model.Moviment{ AccountID: varAcc,
								TransactionAt: convertDate}

	// call service
	res, err := h.workerService.GetAccountMovimentStatementPerDate(req.Context(), moviment)
	if err != nil {
		switch err {
		case erro.ErrNotFound:
			core_apiError = core_apiError.NewAPIError(err, http.StatusNotFound)
		default:
			core_apiError = core_apiError.NewAPIError(err, http.StatusInternalServerError)
		}
		return &core_apiError
	}
	
	return core_json.WriteJSON(rw, http.StatusOK, res)
}