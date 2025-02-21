package config

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"github.com/xyenon/telemikiya/types"

	_ "embed"
)

type Config struct {
	Telegram  Telegram  `mapstructure:"telegram"`
	Database  Database  `mapstructure:"database"`
	Embedding Embedding `mapstructure:"embedding"`
}

type Telegram struct {
	APIID   int    `mapstructure:"api_id"`
	APIHash string `mapstructure:"api_hash"`

	PhoneNumber          string        `mapstructure:"phone_number"`
	UserSessionFile      string        `mapstructure:"user_session_file"`
	ObservedDialogIDs    []int64       `mapstructure:"observed_dialog_ids"`
	DialogUpdateInterval time.Duration `mapstructure:"dialog_update_interval"`

	BotToken          string  `mapstructure:"bot_token"`
	BotSessionFile    string  `mapstructure:"bot_session_file"`
	BotAllowedUserIDs []int64 `mapstructure:"bot_allowed_user_ids"`
}

// Database represents the database configuration.
// https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters
type Database struct {
	Host           string        `mapstructure:"host"`
	Port           uint16        `mapstructure:"port"`
	SSLMode        string        `mapstructure:"ssl_mode"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	User           string        `mapstructure:"user"`
	Password       string        `mapstructure:"password"`
	DBName         string        `mapstructure:"db_name"`
}

type Embedding struct {
	Provider   types.ProviderType `mapstructure:"provider"`
	BaseURL    string             `mapstructure:"base_url"`
	Timeout    time.Duration      `mapstructure:"timeout"`
	BatchSize  uint               `mapstructure:"batch_size"`
	Model      string             `mapstructure:"model"`
	Dimensions uint               `mapstructure:"dimensions"`
	Ollama     Ollama             `mapstructure:"ollama"`
	OpenAI     OpenAI             `mapstructure:"openai"`
}

type Ollama struct {
	KeepAlive       time.Duration  `mapstructure:"keep_alive"`
	ModelParameters map[string]any `mapstructure:"model_parameters"`
}

type OpenAI struct {
	APIKey       string `mapstructure:"api_key"`
	Organization string `mapstructure:"organization"`
	Project      string `mapstructure:"project"`
}

//go:embed config.default.toml
var defaultCfg string

func New(cfgFile string) (cfg *Config, err error) {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
	)

	v.SetConfigType("toml")
	err = v.ReadConfig(strings.NewReader(defaultCfg))
	if err != nil {
		return nil, fmt.Errorf("failed to read default config: %w", err)
	}

	xdgConfigPath := path.Join(xdg.ConfigHome, "telemikiya")
	xdgStatePath := path.Join(xdg.StateHome, "telemikiya")

	v.SetDefault("telegram.user_session_file", path.Join(xdgStatePath, "user_session.db"))
	v.SetDefault("telegram.bot_session_file", path.Join(xdgStatePath, "bot_session.db"))

	if lo.IsNotEmpty(cfgFile) {
		v.SetConfigFile(cfgFile)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath(xdgConfigPath)
		v.AddConfigPath("/etc/telemikiya")
	}

	err = v.MergeInConfig()
	switch err.(type) {
	case nil, viper.ConfigFileNotFoundError:
		// do nothing
	default:
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	v.SetEnvPrefix("TELEMIKIYA")
	v.AutomaticEnv()

	if err = v.Unmarshal(&cfg); err != nil {
		err = fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return
}
