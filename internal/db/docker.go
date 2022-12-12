package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/obalunenko/instadiff-cli/internal/utils"

	log "github.com/obalunenko/logger"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	logPfx = "dockertest: "
	mdb    = "mongo"
)

// ContainerParams holds mongo container parameters.
type ContainerParams struct {
	User          string
	UserPassword  string
	ExpireSeconds uint
}

// SetUpMongoContainer starts up mongoDB docker container and returns reset closure func, that should be run
// after m.Run().
func SetUpMongoContainer(ctx context.Context, m *testing.M, tag string, p ContainerParams) func() {
	log.Info(ctx, logPfx+"Setting up MongoDB test container")

	env := []string{
		// username and password for mongodb superuser
		"MONGO_INITDB_ROOT_USERNAME=" + p.User,
		"MONGO_INITDB_ROOT_PASSWORD=" + p.UserPassword,
	}

	return setUpDB(ctx, m, containerParams{
		repo: mdb,
		tag:  tag,
		port: "27017",
		env:  env,
	}, p)
}

// containerParams holds container request parameter.
type containerParams struct {
	repo string
	tag  string
	port string
	env  []string
}

func setUpDB(ctx context.Context, m *testing.M, container containerParams, p ContainerParams) func() {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.WithError(ctx, err).Fatal(logPfx + "Could not connect to docker")
	}

	if err = pool.Client.Ping(); err != nil {
		log.WithError(ctx, err).Fatal(logPfx + "Could not connect to Docker")
	}

	// pulls a repo, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: container.repo,
		Tag:        container.tag,
		Env:        container.env,
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.WithError(ctx, err).Fatal(logPfx + "Could not start container")
	}

	// Tell docker to hard kill the container in configured expiration time
	if err = resource.Expire(p.ExpireSeconds); err != nil {
		log.WithError(ctx, err).WithFields(log.Fields{
			"container": resource.Container.Name,
		}).Fatal(logPfx + "Could not set expiration to docker container")
	}

	hostport := resource.GetHostPort(fmt.Sprintf("%s/tcp", container.port))

	resetenv := setTestURI(ctx, m, fmt.Sprintf("mongodb://%s:%s@%s", p.User, p.UserPassword, hostport))

	retryFn := func(ctx context.Context) func() error {
		var (
			retries int
			u       string
		)

		u, err = getTestURI()
		if err != nil {
			log.WithError(ctx, err).Fatal(logPfx + "Failed to get test uri")
		}

		return func() error {
			var cl *mongo.Client

			retries++

			log.WithFields(ctx, log.Fields{
				"retries": retries,
				"uri":     u,
			}).Info(logPfx + "Trying to connect to database in container")

			cl, err = mongo.Connect(ctx, options.Client().ApplyURI(u), options.Client().SetDirect(true))
			if err != nil {
				log.WithError(ctx, err).Error(logPfx + "Failed to connect")
				return fmt.Errorf("connect: %w", err)
			}

			defer func() {
				utils.LogError(ctx, cl.Disconnect(ctx), "Failed to disconnect client")
			}()

			if err = cl.Ping(ctx, nil); err != nil {
				log.WithError(ctx, err).Error(logPfx + "Failed to ping")
				return fmt.Errorf("ping: %w", err)
			}

			return nil
		}
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 180 * time.Second
	if err = pool.Retry(retryFn(ctx)); err != nil {
		log.WithError(ctx, err).WithFields(log.Fields{
			"container": resource.Container.Name,
		}).Fatal(logPfx + "Could not connect to docker container")
	}

	log.WithFields(ctx, log.Fields{
		"container": resource.Container.Name,
	}).Info(logPfx + "Container is ready")

	return func() {
		log.WithFields(ctx, log.Fields{
			"container": resource.Container.Name,
		}).Info(logPfx + "Terminating container")

		resetenv()

		// You can't defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.WithError(ctx, err).Fatal(logPfx + "Could not purge resource")
		}

		log.WithFields(ctx, log.Fields{
			"container": resource.Container.Name,
		}).Info(logPfx + "Container is terminated")
	}
}
