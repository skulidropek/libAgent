package utility

import (
	"os"

	"github.com/rs/zerolog/log"
)

func GetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal().Msgf("environment key %s is empty", key)
	}
	return val
}
