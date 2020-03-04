package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

type config struct {
	// MySQL configs.
	User     string
	Password string
	Hostname string
	Port     string
	Database string

	// Debug indicates if in debug mode.
	Debug bool

	// Label is used as prefix in log output, e.g., mainnet, testnet.
	Label string

	// RPCs are backend NEO-CLI nodes used in JSON-RPC queries.
	RPCs []string `mapstructure:"rpcs"`
}

var cfg config

// Load creates a single
func Load(display bool) {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	// Incase test cases require loading configs
	viper.AddConfigPath("../config")

	if err := load(display); err != nil {
		panic(err)
	}
	if err := check(); err != nil {
		panic(err)
	}
}

/* ------------------------------
        `Get` functions
------------------------------ */

// DebugMode tells if running in debug mode.
func DebugMode() bool {
	return cfg.Debug
}

// GetLabel returns custome label as part of the log output prefix.
func GetLabel() string {
	return cfg.Label
}

// GetRPCs returns all rpc urls from config.
func GetRPCs() []string {
	return cfg.RPCs
}

// GetDbConnStr returns db connection string.
func GetDbConnStr() string {
	str := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s",
		cfg.User,
		cfg.Password,
		cfg.Hostname,
		cfg.Port,
		cfg.Database,
	)

	return str
}

/* ------------------------------
         Utility Functions
------------------------------ */

func load(display bool) error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}

	if display {
		configContent, err := json.MarshalIndent(cfg, "", "    ")
		if err != nil {
			panic(err)
		}

		log.Println(string(configContent))
	}

	return nil
}

func check() error {
	if err := checkRPCs(); err != nil {
		return err
	}

	return nil
}

func checkRPCs() error {
	if len(cfg.RPCs) < 1 {
		return errors.New("at least 1 rpc server url must be set")
	}

	for _, rpc := range cfg.RPCs {
		if strings.HasPrefix(rpc, "http") {
			u, err := url.Parse(rpc)
			if err != nil {
				return err
			}
			rpc = u.Host
		}

		_, _, err := net.SplitHostPort(rpc)
		if err != nil {
			return err
		}
	}

	return nil
}
