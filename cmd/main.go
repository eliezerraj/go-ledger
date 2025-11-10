package main

import(
	"time"
	"context"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-ledger/internal/infra/configuration"
	"github.com/go-ledger/internal/core/model"
	"github.com/go-ledger/internal/core/service"
	"github.com/go-ledger/internal/infra/server"
	"github.com/go-ledger/internal/adapter/api"
	"github.com/go-ledger/internal/adapter/database"

	"github.com/go-ledger/internal/adapter/event"
	
	go_core_pg "github.com/eliezerraj/go-core/database/pg"
	go_core_api "github.com/eliezerraj/go-core/api"
)

var(
	logLevel = 	zerolog.InfoLevel // zerolog.InfoLevel zerolog.DebugLevel
	appServer			model.AppServer
	databaseConfig 		go_core_pg.DatabaseConfig
	databasePGServer 	go_core_pg.DatabasePGServer
	childLogger = log.With().Str("component","go-ledger").Str("package", "main").Logger()
)

// Above init
func init(){
	childLogger.Info().Str("func","init").Send()
	zerolog.SetGlobalLevel(logLevel)

	infoPod, server := configuration.GetInfoPod()
	configOTEL 		:= configuration.GetOtelEnv()
	databaseConfig 	:= configuration.GetDatabaseEnv() 
	apiEndpoint	 	:= configuration.GetEndpointEnv() 
	kafkaConfigurations, topics := configuration.GetKafkaEnv() 

	appServer.InfoPod = &infoPod
	appServer.Server = &server
	appServer.ConfigOTEL = &configOTEL
	appServer.DatabaseConfig = &databaseConfig
	appServer.ApiService = apiEndpoint
	appServer.KafkaConfigurations = &kafkaConfigurations
	appServer.Topics = topics
}

// Above main
func main (){
	childLogger.Info().Str("func","main").Interface("appServer",appServer).Send()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Open Database
	count := 1
	var err error
	for {
		databasePGServer, err = databasePGServer.NewDatabasePGServer(ctx, *appServer.DatabaseConfig)
		if err != nil {
			if count < 3 {
				log.Error().Err(err).Msg("error open database... trying again !!")
			} else {
				log.Error().Err(err).Msg("fatal error open Database aborting")
				panic(err)
			}
			time.Sleep(3 * time.Second) //backoff
			count = count + 1
			continue
		}
		break
	}

	// Kafka
	workerEvent, err := event.NewWorkerEventTX(ctx, appServer.Topics, appServer.KafkaConfigurations)
	if err != nil {
		childLogger.Error().Err(err).Send()
	}
	
	// Create a go-core api service for client http
	coreRestApiService := go_core_api.NewRestApiService()

	// wire	
	database := database.NewWorkerRepository(&databasePGServer)
	workerService := service.NewWorkerService(	*coreRestApiService,
												database,
												appServer.ApiService, 
												workerEvent )
	httpRouters := api.NewHttpRouters(workerService, time.Duration(appServer.Server.CtxTimeout))

	// Services Health Check
	err = workerService.HealthCheck(ctx)
	if err != nil {
		log.Error().Err(err).Msg("fatal error health check aborting")
	} else {
		childLogger.Info().Msg("SERVICES HEALTH CHECK OK")
	}

	// start server
	httpServer := server.NewHttpAppServer(appServer.Server)
	httpServer.StartHttpAppServer(ctx, &httpRouters, &appServer)

	// Close kafka producer
	defer func() { 
		workerEvent.WorkerKafka.Close()
		childLogger.Info().Msg("stop main done !!!")
	}()
}