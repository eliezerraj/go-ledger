package event

import (
	"context"

	go_core_observ "github.com/eliezerraj/go-core/observability"
	go_core_event "github.com/eliezerraj/go-core/event/kafka"

	"github.com/rs/zerolog/log"
)

var (
	childLogger = log.With().Str("component","go-ledger").Str("package","internal.adapter.event").Logger()
	tracerProvider go_core_observ.TracerProvider
	producerWorker go_core_event.ProducerWorker
)

type WorkerEvent struct {
	Topics				[]string
	WorkerKafka 		*go_core_event.ProducerWorker 
	KafkaConfigurations *go_core_event.KafkaConfigurations
}

// About create a worker producer kafka
func NewWorkerEvent(ctx context.Context, topics []string, kafkaConfigurations *go_core_event.KafkaConfigurations) (*WorkerEvent, error) {
	childLogger.Info().Str("func","NewWorkerEvent").Send()

	//trace
	span := tracerProvider.Span(ctx, "adapter.event.NewWorkerEvent")
	defer span.End()

	workerKafka, err := producerWorker.NewProducerWorker(kafkaConfigurations)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, err
	}

	return &WorkerEvent{
		Topics: topics,
		WorkerKafka: workerKafka,
		KafkaConfigurations: kafkaConfigurations,
	},nil
}

// About create a worker producer kafka with transaction
func NewWorkerEventTX(ctx context.Context, topics []string, kafkaConfigurations *go_core_event.KafkaConfigurations) (*WorkerEvent, error) {
	childLogger.Info().Str("func","NewWorkerEventTX").Send()

	//trace
	span := tracerProvider.Span(ctx, "adapter.event.NewWorkerEventTX")
	defer span.End()

	workerKafka, err := producerWorker.NewProducerWorkerTX(kafkaConfigurations)
	if err != nil {
		return nil, err
	}

	// Start Kafka InitTransactions
	err = workerKafka.InitTransactions(ctx)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return nil, err
	}

	return &WorkerEvent{
		Topics: topics,
		WorkerKafka: workerKafka,
		KafkaConfigurations: kafkaConfigurations,
	},nil
}

// Above destroy ans create a new producer in case of failed abort
func (w *WorkerEvent) DestroyWorkerEventProducerTx(ctx context.Context) (error) {
	childLogger.Info().Str("func","DestroyWorkerEventProducerTx").Send()

	//trace
	span := tracerProvider.Span(ctx, "adapter.event.DestroyWorkerEventProducerTx")
	defer span.End()

	w.WorkerKafka.Close()	

	new_workerKafka, err := producerWorker.NewProducerWorkerTX(w.KafkaConfigurations)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return err
	}

	// Start Kafka InitTransactions
	err = new_workerKafka.InitTransactions(ctx)
	if err != nil {
		childLogger.Error().Err(err).Send()
		return  err
	}

	w.WorkerKafka = new_workerKafka

	return nil
}