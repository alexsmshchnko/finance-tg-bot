package config

import (
	"log"
	"sync"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfighcl"
)

type Config struct {
	TelegramBotToken  string `hcl:"tg_bot_token" env:"TG_BOT_TOKEN" required:"true"`
	YdbDSN            string `hcl:"ydb_dsn" env:"YDB_DSN" required:"true"`
	SAKeyFileCredPath string `hcl:"sa_path" env:"SA_PATH"`
	ServerPort        string `env:"PORT"`
}

var (
	cfg  Config
	once sync.Once
)

func Get() Config {
	once.Do(func() {
		loader := aconfig.LoaderFor(&cfg, aconfig.Config{
			//EnvPrefix: "",
			Files: []string{"./internal/config/config.local.hcl", "./config.hcl", "$HOME/.config/finance-tg-bot/config.hcl"},
			FileDecoders: map[string]aconfig.FileDecoder{
				".hcl": aconfighcl.New(),
			},
		})

		if err := loader.Load(); err != nil {
			log.Printf("[ERROR] failed to load config: %v", err)
		}
	})

	return cfg
}
