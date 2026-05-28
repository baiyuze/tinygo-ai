package config

import "time"

const (
	DefaultSSID     = "儿童拆迁队"
	DefaultPassword = "zongzong1024"

	SetupSSID = "tinygo-s3-setup"
	APIP      = "192.168.4.1"
	HTTPPort  = 80

	PollTime = 5 * time.Millisecond

	MaxSSIDBytes         = 96
	MaxPasswordBytes     = 128
	MaxAPIKeyBytes       = 1024
	MaxSystemPromptBytes = 1024
)

type Credentials struct {
	SSID         string
	Password     string
	APIKey       string
	SystemPrompt string
}
