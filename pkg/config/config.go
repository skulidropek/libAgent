package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

const EnvPrefixKey = "LIBAGENT_ENV_PREFIX"

type Config struct {
	AIURL   string `env:"AI_URL"`
	AIToken string `env:"AI_TOKEN"`
	Model   string `env:"MODEL"`

	ReWOODisable bool `env:"REWOO_DISABLE"`

	SemanticSearchDisable        bool   `env:"SEMANTIC_SEARCH_DISABLE"`
	SemanticSearchAIURL          string `env:"AI_URL,SEMANTIC_SEARCH_AI_URL"`
	SemanticSearchAIToken        string `env:"AI_TOKEN,SEMANTIC_SEARCH_AI_TOKEN"`
	SemanticSearchDBConnection   string `env:"SEMANTIC_SEARCH_DB_CONNECTION"`
	SemanticSearchEmbeddingModel string `env:"SEMANTIC_SEARCH_EMBEDDING_MODEL"`
	SemanticSearchMaxResults     int    `env:"SEMANTIC_SEARCH_MAX_RESULTS"`

	DDGSearchDisable    bool   `env:"DDG_SEARCH_DISABLE"`
	DDGSearchUserAgent  string `env:"DDG_SEARCH_USER_AGENT"`
	DDGSearchMaxResults int    `env:"DDG_SEARCH_MAX_RESULTS"`

	WebReaderDisable bool `env:"WEB_READER_DISABLE"`

	NmapDisable bool `env:"NMAP_DISABLE"`

	SimpleCMDExecutorDisable bool `env:"SIMPLE_CMD_EXECUTOR_DISABLE"`
}

func NewConfig() (Config, error) {
	cfg := Config{}

	err := godotenv.Load()
	if err != nil {
		log.Warn().Err(err).Msg(".env file load")
	}

	envPrefix := os.Getenv(EnvPrefixKey)
	if envPrefix != "" && strings.HasSuffix(envPrefix, "_") {
		envPrefix = strings.TrimSuffix(envPrefix, "_")
	}

	val := reflect.ValueOf(&cfg).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envValue := ""
		for _, key := range strings.Split(envTag, ",") {
			key = strings.TrimSpace(key)
			if envPrefix != "" {
				key = strings.Join([]string{envPrefix, key}, "_")
			}
			if value, ok := os.LookupEnv(key); ok && value != "" {
				envValue = value
			}
		}
		if envValue == "" {
			continue
		}

		fieldVal := val.Field(i)
		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(envValue)
		case reflect.Int:
			intValue, err := strconv.Atoi(envValue)
			if err != nil {
				return Config{}, fmt.Errorf(
					"failed to parse int for field '%s' from env '%s': %w",
					field.Name, envTag, err,
				)
			}
			fieldVal.SetInt(int64(intValue))
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(strings.ToLower(envValue))
			if err != nil {
				return Config{}, fmt.Errorf(
					"failed to parse bool for field '%s' from env '%s': %w",
					field.Name, envTag, err,
				)
			}
			fieldVal.SetBool(boolValue)
		}
	}

	return cfg, nil
}
