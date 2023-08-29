package gapi

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	db "github.com/uwemakan/simplebank/db/sqlc"
	"github.com/uwemakan/simplebank/token"
	"github.com/uwemakan/simplebank/util"
	"github.com/uwemakan/simplebank/worker"
	"google.golang.org/grpc/metadata"
)

var tokenSymmetricKey = util.RandomString(32)
var accessTokenDuration = time.Minute

func newTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		TokenSymmetricKey: tokenSymmetricKey,
		AccessTokenDuration: accessTokenDuration,
	}

	server, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return server
}

func newContextWithBearerToken(t *testing.T, tokenMaker token.Maker, username string, duration time.Duration) context.Context {
	accessToken, _, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	bearerToken := fmt.Sprintf("%s %s", authorizationBearer, accessToken)
	md := metadata.MD{
		authorizationHeader: []string{bearerToken},
	}
	return metadata.NewIncomingContext(context.Background(), md)
}