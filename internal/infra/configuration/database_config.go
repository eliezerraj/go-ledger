package configuration

import(
	"os"
	"strconv"

	"github.com/joho/godotenv"
	go_core_pg "github.com/eliezerraj/go-core/database/pg"
)

func GetDatabaseEnv() go_core_pg.DatabaseConfig {
	childLogger.Info().Str("func","GetDatabaseEnv").Send()

	err := godotenv.Load(".env")
	if err != nil {
		childLogger.Info().Err(err).Send()
	}
	
	var databaseConfig	go_core_pg.DatabaseConfig

	if os.Getenv("DB_HOST") !=  "" {
		databaseConfig.Host = os.Getenv("DB_HOST")
	}
	if os.Getenv("DB_PORT") !=  "" {
		databaseConfig.Port = os.Getenv("DB_PORT")
	}
	if os.Getenv("DB_NAME") !=  "" {	
		databaseConfig.DatabaseName = os.Getenv("DB_NAME")
	}
	if os.Getenv("DB_SCHEMA") !=  "" {	
		databaseConfig.Schema = os.Getenv("DB_SCHEMA")
	}
	if os.Getenv("DB_DRIVER") !=  "" {	
		databaseConfig.Postgres_Driver = os.Getenv("DB_DRIVER")
	}
	if os.Getenv("DB_MAX_CONNECTION") !=  "" {
		intVar, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNECTION"))
		databaseConfig.DbMax_Connection = intVar
	}

	// Get Database Secrets
	file_user, err := os.ReadFile("/var/pod/secret/username")
	if err != nil {
		childLogger.Error().Err(err).Send()
		os.Exit(3)
	}
	file_pass, err := os.ReadFile("/var/pod/secret/password")
	if err != nil {
		childLogger.Error().Err(err).Send()
		os.Exit(3)
	}
	
	databaseConfig.User = string(file_user)
	databaseConfig.Password = string(file_pass)

	return databaseConfig
}