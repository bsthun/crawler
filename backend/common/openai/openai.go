package openai

import (
	"backend/common/config"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func Init(config *config.Config) *openai.Client {
	client := openai.NewClient(
		option.WithBaseURL(*config.OpenaiBaseUrl),
		option.WithAPIKey(*config.OpenaiApiKey),
		option.WithRequestTimeout(60*time.Second),
	)

	return &client
}
