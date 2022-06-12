package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	log "github.com/obalunenko/logger"
	"github.com/stretchr/testify/require"
)

const (
	dbTestPrefix = "dbtest"
	// EnvDBTestURI holds test database URI. Needs to be in format: mongodb://user:password@address
	EnvDBTestURI = "DBTEST_ISTADIFF"
)

func getTestURI() (string, error) {
	if uri, ok := os.LookupEnv(EnvDBTestURI); ok {
		return uri, nil
	}

	return "", fmt.Errorf("env[%s] not set", EnvDBTestURI)
}

func setTestURI(ctx context.Context, m *testing.M, val string) func() {
	return setEnvForTest(ctx, m, kv{
		name: EnvDBTestURI,
		val:  val,
	})
}

type kv struct {
	name string
	val  string
}

func setEnvForTest(ctx context.Context, _ *testing.M, kv kv) func() {
	old := os.Getenv(kv.name)

	if err := os.Setenv(kv.name, kv.val); err != nil {
		log.WithError(ctx, err).Fatal("failed to set env")
	}

	return func() {
		if err := os.Setenv(kv.name, old); err != nil {
			log.WithError(ctx, err).Fatal("failed to set env")
		}
	}
}

// ConnectForTesting returns a connection to a newly created database
// Test cleanup automatically drops the database and closes underlying connections.
func ConnectForTesting(tb testing.TB, dbname string, collection string) *mongoDB {
	ctx := context.Background()

	u, err := getTestURI()
	require.NoError(tb, err)

	suffix := strings.ToLower(tb.Name())
	if dbname != "" {
		suffix = fmt.Sprintf("%s_%s", suffix, dbname)
	}

	dbName := fmt.Sprintf("%s_%d_%s", dbTestPrefix, time.Now().UnixNano(), suffix)

	cl, err := newMongoDB(ctx, MongoParams{
		URL:        u,
		Database:   dbName,
		Collection: collection,
	})
	require.NoError(tb, err)

	tb.Cleanup(func() {
		defer func() {
			require.NoError(tb, cl.client.Disconnect(ctx))
		}()

		require.NoError(tb, cl.database.Drop(ctx))
	})

	return cl
}
