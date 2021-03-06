package app

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	envstruct "code.cloudfoundry.org/go-envstruct"
)

// Config stores configuration data for the accumulator.
type Config struct {
	UAAAddr        string   `env:"UAA_ADDR,        required"`
	ClientID       string   `env:"CLIENT_ID,       required"`
	ClientSecret   string   `env:"CLIENT_SECRET,   required, noreport"`
	NozzleAddrs    []string `env:"NOZZLE_ADDRS,    required"`
	Port           uint16   `env:"PORT,            required"`
	SkipCertVerify bool     `env:"SKIP_CERT_VERIFY"`

	// RateInterval is used with the rates endpoint. This should match the
	// POLLING_INTERVAL of the nozzle.
	RateInterval time.Duration `env:"RATE_INTERVAL"`

	// VCapApplication is used to detect whether or not the application is
	// deployed as a CF application. If it is the NOZZLE_COUNT and
	// NOZZLE_APP_GUID  are required.
	VCapApplication string `env:"VCAP_APPLICATION"`
	NozzleCount     int    `env:"NOZZLE_COUNT"`
	NozzleAppGUID   string `env:"NOZZLE_APP_GUID"`
	LogWriter       io.Writer

        // Minimum log level for logger (debug, info, error, fatal)
	MinLogLevel     string `env:"MIN_LOG_LEVEL"`

	TLSConfig *tls.Config
}

// LoadConfig loads the configuration settings from the current environment.
func LoadConfig() Config {
	cfg := Config{
		SkipCertVerify: false,
		RateInterval:   time.Minute,
		LogWriter:      os.Stdout,
		MinLogLevel:    "info",
	}

	if err := envstruct.Load(&cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	// If deployed as a CF application, validate additional required
	// configuration and update NozzleAddrs to have same number of addresses as
	// NozzleCount.
	if cfg.VCapApplication != "" {
		cfg.LogWriter = ioutil.Discard

		if cfg.NozzleCount == 0 {
			log.Fatalf("failed to load config: NOZZLE_COUNT must not be 0 when deployed as CF application")
		}

		if len(cfg.NozzleAddrs) != 1 {
			log.Fatalf("failed to load config: NOZZLE_ADDRS must contain only 1 address when deployed as a CF application")
		}

		if cfg.NozzleAppGUID == "" {
			log.Fatalf("failed to load config: NOZZLE_APP_GUID cannot be empty when deployed as CF application")
		}

		addrs := make([]string, 0, cfg.NozzleCount)
		for i := 0; i < cfg.NozzleCount; i++ {
			addrs = append(addrs, cfg.NozzleAddrs[0])
		}

		cfg.NozzleAddrs = addrs
	}

	cfg.TLSConfig = &tls.Config{InsecureSkipVerify: cfg.SkipCertVerify}

	return cfg
}
