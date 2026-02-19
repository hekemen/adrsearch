package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func CreateDatabase(ctx context.Context) (*postgres.PostgresContainer, error) {
	log.Info().Msg("🚀 Starting pgvector container...")
	pgContainer, err := postgres.Run(ctx,
		"pgvector/pgvector:pg16",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(1*time.Minute)),
		testcontainers.WithReuseByName("pgvector"),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			Reuse: true,
		}),
		// testcontainers.WithHostConfigModifier(func(hostConfig *container.HostConfig) {
		// 	hostConfig.Binds = []string{
		// 		"/home/hekemen/oclient/data:/var/lib/postgresql/data",
		// 	}
		// }),
	)
	if err != nil {
		log.Err(err).Msg("failed to start container")
		return nil, fmt.Errorf("")
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Err(err).Msg("failed to start container")
		return nil, fmt.Errorf("")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Err(err).Msg("database")
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer db.Close()

	setupDatabase(db)

	log.Info().Msg("✅ Container is up, extension 'vector' is enabled, and table is ready!")

	return pgContainer, nil
}

func setupDatabase(db *sql.DB) {
	_, err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		log.Err(err).Msg("Failed to create vector extension")
	}

	query := `
	CREATE TABLE IF NOT EXISTS neighborhoods (
		id INT,
		province_id INT,
		district_id INT,
		province TEXT,
		district TEXT,
		name TEXT,
		embedding vector(1024) 
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Err(err).Msg("Failed to create table")
	}
}
