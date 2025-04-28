package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIURL   string `env:"OPENAI_URL"`
	OpenAIToken string `env:"OPENAI_TOKEN"`
	Model       string `env:"MODEL"`

	SemanticSearchOpenAIURL      string `env:"OPENAI_URL,SEMANTIC_SEARCH_OPENAI_URL"`
	SemanticSearchOpenAIToken    string `env:"OPENAI_TOKEN,SEMANTIC_SEARCH_OPENAI_TOKEN"`
	SemanticSearchDBConnection   string `env:"SEMANTIC_SEARCH_DB_CONNECTION"`
	SemanticSearchEmbeddingModel string `env:"SEMANTIC_SEARCH_EMBEDDING_MODEL"`
	SemanticSearchMaxResults     int    `env:"SEMANTIC_SEARCH_MAX_RESULTS"`

	DDGSearchUserAgent  string `env:"DDG_SEARCH_USER_AGENT"`
	DDGSearchMaxResults int    `env:"DDG_SEARCH_MAX_RESULTS"`
}

func NewConfig() (Config, error) {
	cfg := Config{}

	err := godotenv.Load()
	if err != nil {
		return cfg, fmt.Errorf("godotenv failed to load .env: %w", err)
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
			if value, ok := os.LookupEnv(key); ok && value != "" {
				envValue = value
			}
		}

		if envValue != "" {
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
	}

	return cfg, nil
}
