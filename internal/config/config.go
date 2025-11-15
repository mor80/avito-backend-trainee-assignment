package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
)

type (
	Config struct {
		App      App      `koanf:"app"`
		HTTP     HTTP     `koanf:"http"`
		Postgres Postgres `koanf:"postgres"`
	}

	App struct {
		Name     string `koanf:"name"`
		LogLevel string `koanf:"log_level"`
		Env      string `koanf:"env"`
	}

	HTTP struct {
		Host         string        `koanf:"host"`
		Port         int           `koanf:"port"`
		ReadTimeout  time.Duration `koanf:"read_timeout"`
		WriteTimeout time.Duration `koanf:"write_timeout"`
	}

	Postgres struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		User     string `koanf:"user"`
		Password string `koanf:"password"`
		DBName   string `koanf:"db_name"`
		DSN      string `koanf:"dsn"`
	}
)

var (
	httpHostFlag = flag.String("host", "", "override HTTP host")
	httpPortFlag = flag.Int("port", 0, "override HTTP port")
)

func DSN(pg Postgres) string {
	if pg.DSN != "" {
		return pg.DSN
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pg.User,
		pg.Password,
		pg.Host,
		pg.Port,
		pg.DBName,
	)
}

func Load(configPath string) (*Config, error) {
	k := koanf.New(".")

	// default config
	if err := loadDefaults(k); err != nil {
		return nil, fmt.Errorf("config: load default: %w", err)
	}

	if configPath == "" {
		_ = godotenv.Load()
		configPath = os.Getenv("CONFIG_PATH")
		if configPath == "" {
			return nil, fmt.Errorf("CONFIG_PATH is not set")
		}
	}

	// yaml config
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("config: load yaml %q: %w", configPath, err)
	}

	// env config
	if err := k.Load(env.Provider(
		"REVIEWER_",
		"__",
		func(s string) string {
			return strings.ToLower(strings.TrimPrefix(s, "REVIEWER_"))
		}), nil); err != nil {
		return nil, fmt.Errorf("config: load env: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}

	// cli flags
	flag.Parse()

	if *httpHostFlag != "" {
		cfg.HTTP.Host = *httpHostFlag
	}

	if *httpPortFlag != 0 {
		cfg.HTTP.Port = *httpPortFlag
	}

	if cfg.Postgres.DSN == "" {
		cfg.Postgres.DSN = DSN(cfg.Postgres)
	}

	return &cfg, nil
}

func loadDefaults(k *koanf.Koanf) error {
	defaults := Config{
		App: App{
			Name:     "service-reviewer",
			LogLevel: "info",
			Env:      "local",
		},
		HTTP: HTTP{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		Postgres: Postgres{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			DBName:   "service-reviewer",
			DSN:      "",
		},
	}

	return k.Load(structs.Provider(defaults, "koanf"), nil)
}
