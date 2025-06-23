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
	Telegram    Telegram    `mapstructure:"telegram"`
	Database    Database    `mapstructure:"database"`
	LLMProvider LLMProvider `mapstructure:"llm_provider"`
	Embedding   Embedding   `mapstructure:"embedding"`
	ImageToText ImageToText `mapstructure:"image_to_text"`
}

type Telegram struct {
	APIID   int    `mapstructure:"api_id"`
	APIHash string `mapstructure:"api_hash"`

	PhoneNumber          string        `mapstructure:"phone_number"`
	ObservedDialogIDs    []int64       `mapstructure:"observed_dialog_ids"`
	DialogUpdateInterval time.Duration `mapstructure:"dialog_update_interval"`

	BotToken          string  `mapstructure:"bot_token"`
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

type LLMProvider struct {
	Ollama LLMProviderOllama `mapstructure:"ollama"`
	OpenAI LLMProviderOpenAI `mapstructure:"openai"`
	Google LLMProviderGoogle `mapstructure:"google"`
}

func (p LLMProvider) Embedding(provider types.ProviderType) LLMProviderEmbedding {
	switch provider {
	case types.TypeOllama:
		return p.Ollama.Embedding
	case types.TypeOpenAI:
		return p.OpenAI.Embedding
	case types.TypeGoogle:
		return p.Google.Embedding
	default:
		return LLMProviderEmbedding{}
	}
}

type LLMProviderCommon struct {
	BaseURL string        `mapstructure:"base_url"`
	APIKey  string        `mapstructure:"api_key"`
	Timeout time.Duration `mapstructure:"timeout"`
}

type LLMProviderEmbedding struct {
	Model           string         `mapstructure:"model"`
	ModelParameters map[string]any `mapstructure:"model_parameters"`
	Dimensions      uint           `mapstructure:"dimensions"`
}

type LLMProviderImageToText struct {
	Model           string         `mapstructure:"model"`
	ModelParameters map[string]any `mapstructure:"model_parameters"`
}

type LLMProviderOllama struct {
	LLMProviderCommon `mapstructure:",squash"`
	Embedding         LLMProviderEmbedding   `mapstructure:"embedding"`
	ImageToText       LLMProviderImageToText `mapstructure:"image_to_text"`
	KeepAlive         time.Duration          `mapstructure:"keep_alive"`
}

type LLMProviderOpenAI struct {
	LLMProviderCommon `mapstructure:",squash"`
	Embedding         LLMProviderEmbedding   `mapstructure:"embedding"`
	ImageToText       LLMProviderImageToText `mapstructure:"image_to_text"`
	Organization      string                 `mapstructure:"organization"`
	Project           string                 `mapstructure:"project"`
}

type LLMProviderGoogle struct {
	LLMProviderCommon `mapstructure:",squash"`
	Embedding         LLMProviderEmbedding   `mapstructure:"embedding"`
	ImageToText       LLMProviderImageToText `mapstructure:"image_to_text"`
	QuotaProject      string                 `mapstructure:"quota_project"`
}

type Embedding struct {
	Provider  types.ProviderType `mapstructure:"provider"`
	BatchSize uint               `mapstructure:"batch_size"`
}

type ImageToText struct {
	Provider types.ProviderType `mapstructure:"provider"`
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
