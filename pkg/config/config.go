package config

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
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

	CommandExecutorDisable  bool              `env:"COMMAND_EXECUTOR_DISABLE"`
	CommandExecutorCommands map[string]string `env:"COMMAND_EXECUTOR_CMD_*"`
}

func NewConfig() (Config, error) {
	cfg := Config{}

	if err := godotenv.Load(); err != nil {
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

		if strings.HasSuffix(envTag, "_*") {
			if err := processWildcardField(val.Field(i), field, envTag, envPrefix); err != nil {
				return Config{}, err
			}
		} else {
			if err := processSingleField(val.Field(i), field, envTag, envPrefix); err != nil {
				return Config{}, err
			}
		}
	}

	return cfg, nil
}

func processSingleField(fieldVal reflect.Value, field reflect.StructField, envTag, envPrefix string) error {
	envValue := ""
	for _, key := range strings.Split(envTag, ",") {
		key = strings.TrimSpace(key)
		if envPrefix != "" {
			key = strings.Join([]string{envPrefix, key}, "_")
		}
		if value, ok := os.LookupEnv(key); ok && value != "" {
			envValue = value
			break
		}
	}
	if envValue != "" {
		return setFieldValue(fieldVal, envValue, field.Name, envTag)
	}
	return nil
}

func processWildcardField(fieldVal reflect.Value, field reflect.StructField, envTag, envPrefix string) error {
	keyPrefix := strings.TrimSuffix(envTag, "_*")
	if envPrefix != "" {
		keyPrefix = strings.Join([]string{envPrefix, keyPrefix}, "_")
	}

	envMap := map[string]string{}
	envArrayString := map[int]string{}
	var mapKeyType reflect.Kind

	fieldKind := fieldVal.Kind()

	switch fieldKind {
	case reflect.Map:
		mapType := field.Type
		mapKeyType = mapType.Key().Kind()
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.MakeMap(mapType))
		}
	case reflect.Slice:
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.MakeSlice(field.Type, 0, 0))
		}
	default:
		return fmt.Errorf("field '%s' with wildcard tag is not a map or slice", field.Name)
	}

	regex := fmt.Sprintf("^%s_(.+)$", regexp.QuoteMeta(keyPrefix))
	r := regexp.MustCompile(regex)

	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]

		if matches := r.FindStringSubmatch(key); len(matches) == 2 {
			suffix := matches[1]
			switch fieldKind {
			case reflect.Map:
				envMap[suffix] = value
			case reflect.Slice:
				if index, err := strconv.Atoi(suffix); err == nil {
					envArrayString[index] = value
				} else if mapKeyType == reflect.String {
					envMap[suffix] = value
				}
			}
		}
	}

	switch fieldKind {
	case reflect.Map:
		mapKeyType := field.Type.Key().Kind()
		mapValueType := field.Type.Elem()
		for k, v := range envMap {
			mapValue := reflect.New(mapValueType).Elem()
			if err := setFieldValue(mapValue, v, fmt.Sprintf("%s[%s]", fieldVal.Type().String(), k), envTag); err != nil {
				return err
			}
			keyVal := reflect.New(field.Type.Key()).Elem()
			if mapKeyType == reflect.String {
				keyVal.SetString(k)
			} else {
				return fmt.Errorf("unsupported map key type for wildcard '%s': %v", envTag, mapKeyType)
			}
			fieldVal.SetMapIndex(keyVal, mapValue)
		}
		return nil
	case reflect.Slice:
		sliceElemType := field.Type.Elem()
		if len(envArrayString) > 0 {
			sliceVal := reflect.MakeSlice(field.Type, len(envArrayString), len(envArrayString))
			for i, v := range envArrayString {
				elemVal := sliceVal.Index(i)
				if err := setFieldValue(elemVal, v, fmt.Sprintf("%s[%d]", fieldVal.Type().String(), i), envTag); err != nil {
					return err
				}
			}
			fieldVal.Set(sliceVal)
		} else if len(envMap) > 0 {
			sliceVal := reflect.MakeSlice(field.Type, 0, len(envMap))
			for _, v := range envMap {
				elemVal := reflect.New(sliceElemType).Elem()
				if err := setFieldValue(elemVal, v, fmt.Sprintf("%s[]", fieldVal.Type().String()), envTag); err != nil {
					return err
				}
				sliceVal = reflect.Append(sliceVal, elemVal)
			}
			fieldVal.Set(sliceVal)
		}
		return nil
	}

	return nil
}

func setFieldValue(fieldVal reflect.Value, envValue string, fieldName, envTag string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(envValue)
	case reflect.Int:
		intValue, err := strconv.Atoi(envValue)
		if err != nil {
			return fmt.Errorf("failed to parse int for field '%s' from env '%s': %w", fieldName, envTag, err)
		}
		fieldVal.SetInt(int64(intValue))
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(strings.ToLower(envValue))
		if err != nil {
			return fmt.Errorf("failed to parse bool for field '%s' from env '%s': %w", fieldName, envTag, err)
		}
		fieldVal.SetBool(boolValue)
	default:
		return fmt.Errorf("unsupported type for field '%s' with env tag '%s'", fieldName, envTag)
	}
	return nil
}
