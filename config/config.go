package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/stenstromen/ncregistry/types"
)

var Config types.Config

func SaveConfig(newEntry types.Entry) {
	Config.Entries = append(Config.Entries, newEntry)
	viper.Set("Entries", Config.Entries)
	if err := viper.WriteConfig(); err != nil {
		log.Fatalf("Error writing config: %s", err)
	}
}

func InitConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %s", err)
	}

	configDir := filepath.Join(home, ".ncregistry")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.Mkdir(configDir, 0700); err != nil {
			log.Fatalf("Failed to create directory: %s", err)
		}
	}

	configFile := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		file, err := os.Create(configFile)
		if err != nil {
			log.Fatalf("Failed to create file: %s", err)
		}
		file.Close()
	}

	viper.SetConfigFile(configFile)

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.SafeWriteConfig()
		} else {
			fmt.Printf("Error reading config file: %s", err)
		}
	}

	if err := viper.Unmarshal(&Config); err != nil {
		fmt.Printf("Unable to decode into struct, %v", err)
	}
}
