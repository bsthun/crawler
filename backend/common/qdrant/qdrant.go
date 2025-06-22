package qdrant

import (
	"backend/common/config"
	"github.com/bsthun/gut"
	"github.com/qdrant/go-client/qdrant"
	"strconv"
	"strings"
)

func Init(config *config.Config) *qdrant.Client {
	segments := strings.Split(*config.QdrantDsn, ":")
	host := segments[0]
	port, err := strconv.ParseInt(segments[1], 10, 64)

	// * create qdrant client
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:                   host,
		Port:                   int(port),
		APIKey:                 "",
		UseTLS:                 false,
		TLSConfig:              nil,
		GrpcOptions:            nil,
		SkipCompatibilityCheck: false,
	})
	if err != nil {
		gut.Fatal("unable to create qdrant client", err)
	}

	return client
}
