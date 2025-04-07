package configuration

import(
	"os"
	"strings"
	
	"github.com/joho/godotenv"
	go_core_observ "github.com/eliezerraj/go-core/observability" 
)

func GetOtelEnv() go_core_observ.ConfigOTEL {
	childLogger.Info().Str("func","GetOtelEnv").Send()

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Send()
	}

	var configOTEL	go_core_observ.ConfigOTEL

	configOTEL.TimeInterval = 1
	configOTEL.TimeAliveIncrementer = 1
	configOTEL.TotalHeapSizeUpperBound = 100
	configOTEL.ThreadsActiveUpperBound = 10
	configOTEL.CpuUsageUpperBound = 100
	configOTEL.SampleAppPorts = []string{}

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") !=  "" {	
		configOTEL.OtelExportEndpoint = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	if os.Getenv("USE_STDOUT_TRACER_EXPORTER") ==  "true" {
		configOTEL.UseStdoutTracerExporter = true
	} else {
		configOTEL.UseStdoutTracerExporter = false
	}

	if os.Getenv("USE_OTLP_COLLECTOR") ==  "true" {
		configOTEL.UseOtlpCollector = true
	} else {
		configOTEL.UseOtlpCollector = false
	}

	if os.Getenv("AWS_CLOUDWATCH_LOG_GROUP") !=  "" {	
		configOTEL.AWSCloudWatchLogGroup = strings.Split(os.Getenv("AWS_CLOUDWATCH_LOG_GROUP"),",")
	}

	return configOTEL
}