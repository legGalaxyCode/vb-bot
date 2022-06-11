package config

import (
	"fmt"
	_ "github.com/golang/protobuf/ptypes/any"
	"github.com/joho/godotenv"
	"os"
	"sync"
)

type BotConfig struct {
	Token        string `env:"TOKEN"`
	BotMsgPrefix string `env:"BOT_PREFIX"`
}

type SheetsConfig struct {
	AccessToken string `env:"ACCESS_TOKEN"`
}

type DrivePermissionConfig struct {
	Type string `env:"USER_TYPE"`
	Role string `env:"USER_ROLE"`
}

type FilepathConfig struct {
	DownloadsPath        string `env:"DOWNLOADS_PATH"`
	SpreadsheetDataPath  string `env:"SPREADSHEETDATA_PATH"`
	SpreadsheetPropsPath string `env:"SPREADSHEETPROPS_PATH"`
	ClientPath           string `env:"CLIENT_PATH"`
	ClientCreditsPath    string `env:"CLIENTCREDITS_PATH"`
	TokenPath            string `env:"TOKEN_PATH"`
}

type Config struct {
	Bot      BotConfig
	Sheet    SheetsConfig
	Drive    DrivePermissionConfig
	Filepath FilepathConfig
}

var cfg *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		if err := godotenv.Load("config.env"); err != nil {
			fmt.Println(err)
		}
		cfg = &Config{
			Bot: BotConfig{
				Token:        os.Getenv("TOKEN"),
				BotMsgPrefix: os.Getenv("BOT_PREFIX"),
			},
			Sheet: SheetsConfig{
				AccessToken: os.Getenv("ACCESS_TOKEN"),
			},
			Drive: DrivePermissionConfig{
				Role: os.Getenv("ROLE"),
				Type: os.Getenv("TYPE"),
			},
			Filepath: FilepathConfig{
				DownloadsPath:        os.Getenv("DOWNLOADS_PATH"),
				SpreadsheetDataPath:  os.Getenv("SPREADSHEETDATA_PATH"),
				SpreadsheetPropsPath: os.Getenv("SPREADSHEETPROPS_PATH"),
				ClientPath:           os.Getenv("CLIENT_PATH"),
				ClientCreditsPath:    os.Getenv("CLIENTCREDITS_PATH"),
				TokenPath:            os.Getenv("TOKEN_PATH"),
			},
		}
	})
	return cfg
}
