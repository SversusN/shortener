package grpcsrv

import (
	"context"
	"errors"
	"github.com/SversusN/shortener/config"
	"github.com/SversusN/shortener/internal/storage/primitivestorage"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"sync"
	"testing"
)

func TestCreateServer(t *testing.T) {
	storage := primitivestorage.NewStorage(nil, errors.New("filename is empty, no store tempdb"))
	c := context.Background()
	cfg := config.Config{
		TrustedSubnet: "",
		GRPCAddress:   "3020",
	}
	wg := &sync.WaitGroup{}
	server := NewGRPCServer(&c, storage, &cfg, wg)
	assert.IsType(t, (*grpc.Server)(nil), server)
}
