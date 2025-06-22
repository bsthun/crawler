package ollama

import (
	"backend/common/config"
	"github.com/bsthun/gut"
	"github.com/ollama/ollama/api"
	"net/http"
	"net/url"
	"time"
)

func Init(config *config.Config) (*api.Client, error) {
	baseUrl, err := url.Parse(*config.OllamaBaseUrl)
	if err != nil {
		gut.Fatal("failed to parse url", err)
	}

	httpClient := &http.Client{
		Timeout: 60 * time.Second,
	}

	client := api.NewClient(baseUrl, httpClient)

	return client, nil
}
