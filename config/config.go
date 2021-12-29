package config

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/hermeznetwork/hermez-core/aggregator"
	"github.com/hermeznetwork/hermez-core/db"
	"github.com/hermeznetwork/hermez-core/etherman"
	"github.com/hermeznetwork/hermez-core/jsonrpc"
	"github.com/hermeznetwork/hermez-core/log"
	"github.com/hermeznetwork/hermez-core/proverclient"
	"github.com/hermeznetwork/hermez-core/sequencer"
	"github.com/hermeznetwork/hermez-core/synchronizer"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config represents the configuration of the entire Hermez Node
type Config struct {
	Log           log.Config
	Database      db.Config
	Etherman      etherman.Config
	RPC           jsonrpc.Config
	Synchronizer  synchronizer.Config
	Sequencer     sequencer.Config
	Aggregator    aggregator.Config
	Prover        proverclient.Config
	NetworkConfig NetworkConfig
}

// Load loads the configuration
func Load(configFilePath string, network string) (*Config, error) {
	var cfg Config
	viper.SetConfigType("toml")

	err := viper.ReadConfig(bytes.NewBuffer([]byte(DefaultValues)))
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return nil, err
	}
	if configFilePath != "" {
		path, fullFile := filepath.Split(configFilePath)

		file := strings.Split(fullFile, ".")

		viper.AddConfigPath(path)
		viper.SetConfigName(file[0])
		viper.SetConfigType(file[1])
	}
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("HERMEZCORE")
	err = viper.ReadInConfig()
	if err != nil {
		_, ok := err.(viper.ConfigFileNotFoundError)
		if ok {
			log.Infof("config file not found")
		} else {
			log.Infof("error reading config file: ", err)
			return nil, err
		}
	}

	err = viper.Unmarshal(&cfg, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc()))
	if err != nil {
		return nil, err
	}
	// Load genesis parameters
	cfg.loadNetworkConfig(network)

	cfgJSON, _ := json.MarshalIndent(cfg, "", "  ")
	log.Infof("Configuration loaded: \n%s\n", string(cfgJSON))
	return &cfg, nil
}