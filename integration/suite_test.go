package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *sqlx.DB
}

func (s *PostgresTestSuite) SetupSuite() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "postgres:15-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.container = container

	port, _ := container.MappedPort(ctx, "5432")
	dsn := fmt.Sprintf("host=localhost port=%s user=test password=test dbname=testdb sslmode=disable", port.Port())
	db, err := sqlx.Connect("postgres", dsn)
	s.Require().NoError(err)
	s.db = db

	// Run migrations
	m, err := migrate.New(
		"file://../migrations",
		dsn,
	)
	s.Require().NoError(err)
	s.Require().NoError(m.Up())

}

func (s *PostgresTestSuite) TearDownSuite() {
	s.db.Close()
	s.container.Terminate(context.Background())
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}
