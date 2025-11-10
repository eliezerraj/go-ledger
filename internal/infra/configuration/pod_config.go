package configuration

import(
	"os"
	"strconv"
	"net"
	"context"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/config"

	"github.com/go-ledger/internal/core/model"
)

var childLogger = log.With().Str("component","go-ledger").Str("package","internal.infra.configuration").Logger()

// Load the Pod configuration
func GetInfoPod() (	model.InfoPod, model.Server) {
	childLogger.Info().Str("func","GetInfoPod").Send()

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Send()
	}

	var infoPod 	model.InfoPod
	var server		model.Server

	if os.Getenv("API_VERSION") !=  "" {
		infoPod.ApiVersion = os.Getenv("API_VERSION")
	}
	if os.Getenv("POD_NAME") !=  "" {
		infoPod.PodName = os.Getenv("POD_NAME")
	}
	if os.Getenv("SETPOD_AZ") == "false" {	
		infoPod.IsAZ = false
	} else {
		infoPod.IsAZ = true
	}
	if os.Getenv("ENV") !=  "" {	
		infoPod.Env = os.Getenv("ENV")
	}
	if os.Getenv("ACCOUNT_ID") !=  "" {	
		infoPod.AccountID = os.Getenv("ACCOUNT_ID")
	}
		
	// Get IP
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Error().Err(err).Send()
		os.Exit(3)
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				infoPod.IPAddress = ipnet.IP.String()
			}
		}
	}
	infoPod.OSPID = strconv.Itoa(os.Getpid())

	// Get AZ only if localtest is true
	if (infoPod.IsAZ) {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			childLogger.Error().Err(err).Send()
			os.Exit(3)
		}
		client := imds.NewFromConfig(cfg)
		response, err := client.GetInstanceIdentityDocument(context.TODO(), &imds.GetInstanceIdentityDocumentInput{})
		if err != nil {
			childLogger.Error().Err(err).Send()
			os.Exit(3)
		}
		infoPod.AvailabilityZone = response.AvailabilityZone	
	} else {
		infoPod.AvailabilityZone = "-"
	}

	if os.Getenv("PORT") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("PORT"))
		server.Port = intVar
	}

	server.ReadTimeout = 60
	server.WriteTimeout = 60
	server.IdleTimeout = 60
	server.CtxTimeout = 5

	if os.Getenv("CTX_TIMEOUT") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("CTX_TIMEOUT"))
		server.CtxTimeout = intVar
	}

	return infoPod, server
}
