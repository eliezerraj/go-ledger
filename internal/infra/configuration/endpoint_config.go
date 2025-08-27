package configuration

import(
	"os"
	"time"
	"strconv"
	"github.com/joho/godotenv"
	"github.com/go-ledger/internal/core/model"
)

// About get services endpoints env var
func GetEndpointEnv() []model.ApiService {
	childLogger.Info().Str("func","GetEndpointEnv").Send()

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Send()
	}
	
	var apiService []model.ApiService

	var apiService00 model.ApiService
	if os.Getenv("URL_SERVICE_00") !=  "" {
		apiService00.Url = os.Getenv("URL_SERVICE_00")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_00") !=  "" {
		apiService00.XApigwApiId = os.Getenv("X_APIGW_API_ID_SERVICE_00")
	}
	if os.Getenv("METHOD_SERVICE_00") !=  "" {
		apiService00.Method = os.Getenv("METHOD_SERVICE_00")
	}
	if os.Getenv("NAME_SERVICE_00") !=  "" {
		apiService00.Name = os.Getenv("NAME_SERVICE_00")
	}
	if os.Getenv("HOST_SERVICE_00") !=  "" {
		apiService00.HostName = os.Getenv("HOST_SERVICE_00")
	}
	if os.Getenv("CLIENT_HTTP_TIMEOUT_00") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("CLIENT_HTTP_TIMEOUT_00"))	
		apiService00.HttpTimeout = time.Duration(intVar) * time.Second
	} else {
		apiService00.HttpTimeout =(5 * time.Second) // default
	}
	apiService = append(apiService, apiService00)

	var apiService01 model.ApiService
	if os.Getenv("URL_SERVICE_01") !=  "" {
		apiService01.Url = os.Getenv("URL_SERVICE_01")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_01") !=  "" {
		apiService01.XApigwApiId = os.Getenv("X_APIGW_API_ID_SERVICE_01")
	}
	if os.Getenv("METHOD_SERVICE_01") !=  "" {
		apiService01.Method = os.Getenv("METHOD_SERVICE_01")
	}
	if os.Getenv("NAME_SERVICE_01") !=  "" {
		apiService01.Name = os.Getenv("NAME_SERVICE_01")
	}
	if os.Getenv("HOST_SERVICE_01") !=  "" {
		apiService01.HostName = os.Getenv("HOST_SERVICE_01")
	}
	if os.Getenv("CLIENT_HTTP_TIMEOUT_01") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("CLIENT_HTTP_TIMEOUT_01"))	
		apiService01.HttpTimeout = time.Duration(intVar) * time.Second
	} else {
		apiService01.HttpTimeout =(5 * time.Second) // default
	}
	apiService = append(apiService, apiService01)

	var apiService02 model.ApiService
	if os.Getenv("URL_SERVICE_02") !=  "" {
		apiService02.Url = os.Getenv("URL_SERVICE_02")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_02") !=  "" {
		apiService02.XApigwApiId = os.Getenv("X_APIGW_API_ID_SERVICE_02")
	}
	if os.Getenv("METHOD_SERVICE_02") !=  "" {
		apiService02.Method = os.Getenv("METHOD_SERVICE_02")
	}
	if os.Getenv("NAME_SERVICE_02") !=  "" {
		apiService02.Name = os.Getenv("NAME_SERVICE_02")
	}
	if os.Getenv("HOST_SERVICE_02") !=  "" {
		apiService02.HostName = os.Getenv("HOST_SERVICE_02")
	}
	if os.Getenv("CLIENT_HTTP_TIMEOUT_02") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("CLIENT_HTTP_TIMEOUT_02"))	
		apiService02.HttpTimeout = time.Duration(intVar) * time.Second
	} else {
		apiService02.HttpTimeout =(5 * time.Second) // default
	}
	apiService = append(apiService, apiService02)

	var apiService03 model.ApiService
	if os.Getenv("URL_SERVICE_03") !=  "" {
		apiService03.Url = os.Getenv("URL_SERVICE_03")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_03") !=  "" {
		apiService03.XApigwApiId = os.Getenv("X_APIGW_API_ID_SERVICE_03")
	}
	if os.Getenv("METHOD_SERVICE_03") !=  "" {
		apiService03.Method = os.Getenv("METHOD_SERVICE_03")
	}
	if os.Getenv("NAME_SERVICE_03") !=  "" {
		apiService03.Name = os.Getenv("NAME_SERVICE_03")
	}
	if os.Getenv("HOST_SERVICE_03") !=  "" {
		apiService03.HostName = os.Getenv("HOST_SERVICE_03")
	}
	if os.Getenv("CLIENT_HTTP_TIMEOUT_03") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("CLIENT_HTTP_TIMEOUT_03"))	
		apiService03.HttpTimeout = time.Duration(intVar) * time.Second
	} else {
		apiService03.HttpTimeout =(5 * time.Second) // default
	}
	apiService = append(apiService, apiService03)

	return apiService
}