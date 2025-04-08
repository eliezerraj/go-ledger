package configuration

import(
	"os"
	"github.com/joho/godotenv"
	"github.com/go-ledger/internal/core/model"
)

// About get services endpoints env var
func GetEndpointEnv() []model.ApiService {
	childLogger.Info().Str("func","GetEndpointEnv").Send()

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Error().Err(err).Send()
	}
	
	var apiService []model.ApiService

	var apiService01 model.ApiService
	if os.Getenv("URL_SERVICE_01") !=  "" {
		apiService01.Url = os.Getenv("URL_SERVICE_01")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_01") !=  "" {
		apiService01.Header_x_apigw_api_id = os.Getenv("X_APIGW_API_ID_SERVICE_01")
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
	apiService = append(apiService, apiService01)

	var apiService02 model.ApiService
	if os.Getenv("URL_SERVICE_02") !=  "" {
		apiService02.Url = os.Getenv("URL_SERVICE_02")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_02") !=  "" {
		apiService02.Header_x_apigw_api_id = os.Getenv("X_APIGW_API_ID_SERVICE_02")
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
	apiService = append(apiService, apiService02)

	var apiService03 model.ApiService
	if os.Getenv("URL_SERVICE_03") !=  "" {
		apiService03.Url = os.Getenv("URL_SERVICE_03")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_03") !=  "" {
		apiService03.Header_x_apigw_api_id = os.Getenv("X_APIGW_API_ID_SERVICE_03")
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
	apiService = append(apiService, apiService03)

	var apiService04 model.ApiService
	if os.Getenv("URL_SERVICE_04") !=  "" {
		apiService04.Url = os.Getenv("URL_SERVICE_04")
	}
	if os.Getenv("X_APIGW_API_ID_SERVICE_04") !=  "" {
		apiService04.Header_x_apigw_api_id = os.Getenv("X_APIGW_API_ID_SERVICE_04")
	}
	if os.Getenv("METHOD_SERVICE_04") !=  "" {
		apiService04.Method = os.Getenv("METHOD_SERVICE_04")
	}
	if os.Getenv("NAME_SERVICE_04") !=  "" {
		apiService04.Name = os.Getenv("NAME_SERVICE_04")
	}
	if os.Getenv("HOST_SERVICE_04") !=  "" {
		apiService03.HostName = os.Getenv("HOST_SERVICE_04")
	}
	apiService = append(apiService, apiService04)

	return apiService
}