package config

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Env struct {
	Port                    string        `mapstructure:"PORT"`
	DBUrl                   string        `mapstructure:"DB_URL"`
	JwtAccessTokenSecretKey string        `mapstructure:"JWT_ACCESS_TOKEN_SECRET_KEY"`
	JwtAccessTokenExpiresIn time.Duration `mapstructure:"JWT_ACCESS_TOKEN_EXPIRES_IN"`
	MailHost                string        `mapstructure:"MAIL_HOST"`
	MailPort                int           `mapstructure:"MAIL_PORT"`
	MailUser                string        `mapstructure:"MAIL_USER"`
	MailPassword            string        `mapstructure:"MAIL_PASSWORD"`
}

func newEnv() (*Env, error) {
	env := &Env{}

	// Cấu hình viper
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.Wrap(err, "failed to read config file")
		}
	}

	// Unmarshal với Decode Hook từ mapstructure
	err := viper.Unmarshal(env, viper.DecodeHook(
		// Kết hợp nhiều hook lại với nhau bằng một hàm tổng hợp
		mapstructure.ComposeDecodeHookFunc(
			stringToJSONHookFunc(),
			stringToTimeDurationHookFunc(),
		),
	))

	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	// Thiết lập các giá trị mặc định
	if env.Port == "" {
		env.Port = "3000"
	}

	return env, nil
}

// stringToTimeDurationHookFunc parse một chuỗi thành time.Duration.
func stringToTimeDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String || t != reflect.TypeOf(time.Duration(0)) {
			return data, nil
		}
		return time.ParseDuration(data.(string))
	}
}

// stringToJSONHookFunc parse một chuỗi JSON thành struct, slice, hoặc map.
func stringToJSONHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		raw := data.(string)
		if raw == "" || (raw[0] != '{' && raw[0] != '[') {
			return data, nil
		}

		switch t.Kind() {
		case reflect.Struct, reflect.Slice, reflect.Map:
			// Hợp lệ, tiếp tục
		default:
			return data, nil
		}

		newVal := reflect.New(t).Interface()
		err := json.Unmarshal([]byte(raw), &newVal)
		if err != nil {
			return data, nil
		}

		return reflect.ValueOf(newVal).Elem().Interface(), nil
	}
}
