package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	db "github.com/uwemakan/simplebank/db/sqlc"
	"github.com/uwemakan/simplebank/util"
)

var tokenSymmetricKey = util.RandomString(32)
var accessTokenDuration = time.Minute

func newTestServer(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey: tokenSymmetricKey,
		AccessTokenDuration: accessTokenDuration,
	}

	server, err := NewServer(config, store)
	require.NoError(t, err)

	return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
