package model

import (
	"time"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_event "github.com/eliezerraj/go-core/event/kafka" 
)

type AppServer struct {
	InfoPod 		*InfoPod 					`json:"info_pod"`
	Server     		*Server     				`json:"server"`
	ConfigOTEL		*go_core_observ.ConfigOTEL	`json:"otel_config"`
	DatabaseConfig	*go_core_pg.DatabaseConfig  `json:"database"`
	ApiService 		[]ApiService				`json:"api_endpoints"`
	KafkaConfigurations							*go_core_event.KafkaConfigurations  `json:"kafka_configurations"`
	Topics 			[]string					`json:"topics"`		
}

type InfoPod struct {
	PodName				string 	`json:"pod_name"`
	ApiVersion			string 	`json:"version"`
	OSPID				string 	`json:"os_pid"`
	IPAddress			string 	`json:"ip_address"`
	AvailabilityZone 	string 	`json:"availabilityZone"`
	IsAZ				bool   	`json:"is_az"`
	Env					string `json:"enviroment,omitempty"`
	AccountID			string `json:"account_id,omitempty"`
}

type Server struct {
	Port 			int `json:"port"`
	ReadTimeout		int `json:"readTimeout"`
	WriteTimeout	int `json:"writeTimeout"`
	IdleTimeout		int `json:"idleTimeout"`
	CtxTimeout		int `json:"ctxTimeout"`
}

type ApiService struct {
	Name			string `json:"name_service"`
	Url				string `json:"url"`
	Method			string `json:"method"`
	XApigwApiId		string `json:"x-apigw-api-id,omitempty"`
	HostName		string `json:"host_name"`
	HttpTimeout		time.Duration `json:"httpTimeout"`
}

type MessageRouter struct {
	Message			string `json:"message"`
}

type Account struct {
	ID				int			`json:"id,omitempty"`
	AccountID		string		`json:"account_id,omitempty"`
	PersonID		string  	`json:"person_id,omitempty"`
	CreatedAt		*time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`
	UserLastUpdate	*string  	`json:"user_last_update,omitempty"`
	TraceID			string		`json:"trace_request_id,omitempty"`
	TenantId		string  	`json:"tenant_id,omitempty"`
}

type Moviment struct {
	AccountFrom		Account  	`json:"account_from,omitempty"`
	AccountTo		*Account  	`json:"account_to,omitempty"`
	TransactionAt	*time.Time 	`json:"transaction_at,omitempty"`
	Type			string  	`json:"type,omitempty"`
	Currency		string  	`json:"currency,omitempty"`
	Amount			float64		`json:"amount,omitempty"`
	TraceID			string		`json:"trace_request_id,omitempty"`
	TenantId		string  	`json:"tenant_id,omitempty"`
}

type MovimentStatement struct {
	MovimentSummary		*[]Moviment	`json:"moviment_summary,omitempty"`
	MovimentStatement	*[]Moviment	`json:"moviment_detail,omitempty"`
}

type MovimentTransaction struct {
	Status				string				`json:"status,omitempty"`
	Transaction			Transaction			`json:"transacion,omitempty"`
	TransactionDetail	[]TransactionDetail	`json:"transacion_detail,omitempty"`
}

type TransactionType struct {
	TransactionTypeID	string		`json:"transaction_type_id,omitempty"`
	Description			string		`json:"description,omitempty"`
}

type Transaction struct {
	ID					int			`json:"id,omitempty"`
	Description			string		`json:"description,omitempty"`
	Currency			string		`json:"fk_ledger_id,omitempty"`
	TransactionAt		time.Time 	`json:"transaction_at,omitempty"`
	UpdatedAt			*time.Time 	`json:"updated_at,omitempty"`
	TraceID				string		`json:"trace_request_id,omitempty"`
	TenantId			string  	`json:"tenant_id,omitempty"`
}

type TransactionDetail struct {
	FkTxID				int			`json:"fk_transaction_id,omitempty"`
	ID					int			`json:"id,omitempty"`
	TransactionAt		time.Time 	`json:"transaction_at,omitempty"`
	FkLedgerID			string		`json:"fk_ledger_id,omitempty"`
	FkAccountID			int			`json:"fk_account_id,omitempty"`
	AccountID			string		`json:"account_id,omitempty"`	
	Status				string  	`json:"status,omitempty"`
	DebitAmount			int64	  	`json:"debit_amount"`
	CreditAmount		int64	  	`json:"credit_amount"`
	CreatedAt			time.Time 	`json:"created_at,omitempty"`
	UpdatedAt			*time.Time 	`json:"updated_at,omitempty"`
	TraceID				string		`json:"trace_request_id,omitempty"`
	TenantId			string  	`json:"tenant_id,omitempty"`
}