package config

import (
	"flag"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	HTTPServer     HTTPServer `env:"database_dsn"`
	DatabaseDSN    string     `env:"DATABASE_URI"`
	SecretKey      []byte
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	WorkerConfig   WorkerConfig
}

type HTTPServer struct {
	Address     string        `env:"RUN_ADDRESS" envDefault:":8080"`
	Timeout     time.Duration `env:"timeout" envDefault:"5s"`
	IdleTimeout time.Duration `env:"idle_timeout" envDefault:"60s"`
}

type WorkerConfig struct {
	CountWorkers   int32
	InputChainSize int
	BufferSize     int
	BatchSize      int
}

func (c *Config) LoadConfigEnv() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	if err := env.Parse(c.HTTPServer.Address); err != nil {
		return err
	}
	if err := env.Parse(c.DatabaseDSN); err != nil {
		return err
	}
	if err := env.Parse(c.AccrualAddress); err != nil {
		return err
	}
	return nil
}
func (c *Config) LoadConfigFlag() {
	if c.HTTPServer.Address == "" {
		flag.StringVar(&c.HTTPServer.Address, "a", ":8080", "HTTP server startup address")
	}
	if c.DatabaseDSN == "" {
		flag.StringVar(&c.DatabaseDSN, "d", "", "DataBase connection string")
	}
	if c.AccrualAddress == "" {
		flag.StringVar(&c.AccrualAddress, "r", "", "ACCRUAL SYSTEM ADDRESS")
	}
	flag.Parse()
}

func MustLoadConfig() *Config {
	cfg := Config{
		HTTPServer: HTTPServer{
			Address: os.Getenv("RUN_ADDRESS"),
		},
		DatabaseDSN:    os.Getenv("DATABASE_DSN"),
		AccrualAddress: os.Getenv("ACCRUAL_SYSTEM_ADDRESS"),
		SecretKey:      []byte("sDfmldsnflkm<M SAD !2scxzcx#454556%$^%^&%*"),
		WorkerConfig: WorkerConfig{
			CountWorkers:   3,
			InputChainSize: 20,
			BufferSize:     10,
			BatchSize:      10,
		},
	}

	cfg.LoadConfigFlag()
	return &cfg
}
