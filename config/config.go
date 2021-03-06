package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"neo3-squirrel/util/convert"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

// MaxVal defines the maximum db decimal storage limitation.
var MaxVal *big.Float

type config struct {
	// MySQL configs.
	User     string
	Password string
	Hostname string
	Port     string
	Database string

	// Debug indicates if in debug mode.
	Debug    bool
	DebugSQL bool

	// Label is used as prefix in log output, e.g., mainnet, testnet.
	Label string

	// RPCs are backend NEO-CLI nodes used in JSON-RPC queries.
	RPCs []string `mapstructure:"rpcs"`

	// Workers sets the number of goroutines that will be created for data processing.
	Workers int
}

var cfg config

// Load creates a single
func Load(display bool, debug, debugSQL bool) {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	// Incase test cases require loading configs
	viper.AddConfigPath("../config")

	if err := load(); err != nil {
		panic(err)
	}

	cfg.Debug = debug
	cfg.DebugSQL = debugSQL

	attachRPCHTTPScheme()

	if err := validateConfig(); err != nil {
		panic(err)
	}

	if display {
		printConfig()
	}

	MaxVal = convert.ToDecimal(fmt.Sprintf("%s.%s", strings.Repeat("9", 35), strings.Repeat("9", 30)))
}

/* ------------------------------
        `Get` functions
------------------------------ */

// DebugMode tells if running in debug mode.
func DebugMode() bool {
	return cfg.Debug
}

// DebugSQLMode tells if shows sql statement.
func DebugSQLMode() bool {
	return cfg.DebugSQL
}

// GetLabel returns custome label as part of the log output prefix.
func GetLabel() string {
	return cfg.Label
}

// GetRPCs returns all rpc urls from config.
func GetRPCs() []string {
	return cfg.RPCs
}

// GetWorkers returns the number of working goroutines.
func GetWorkers() int {
	return cfg.Workers
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

// GetDBInfo returns the connecting DB info.
func GetDBInfo() string {
	return fmt.Sprintf("(%s:%s)/%s", cfg.Hostname, cfg.Port, cfg.Database)
}

/* ------------------------------
         Utility Functions
------------------------------ */

func load() error {
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func printConfig() {
	dbPass := cfg.Password
	if len(dbPass) != 0 {
		cfg.Password = "******"
	}

	configContent, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		panic(err)
	}

	log.Println(string(configContent))
	cfg.Password = dbPass
}

func attachRPCHTTPScheme() {
	for i := 0; i < len(cfg.RPCs); i++ {
		rpc := cfg.RPCs[i]
		if !strings.HasPrefix(rpc, "http") {
			cfg.RPCs[i] = "http://" + rpc
		}
	}
}

func validateConfig() error {
	if err := checkRPCs(); err != nil {
		return err
	}

	if cfg.Workers == 0 {
		return errors.New("workers must be great than 0")
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
