package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hekemen/adrsearch/service"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func main() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	vGenerator, err := service.NewVectorGenerator()
	if err != nil {
		log.Err(err).Msg("creating vector generator")
		return
	}

	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	nChan := make(chan service.Neighborhood)
	grp, mainCtx := errgroup.WithContext(ctx)

	pgContainer, err := service.CreateDatabase(mainCtx)
	if err != nil {
		log.Err(err).Msg("creating database")
		return
	}

	connStr, err := pgContainer.ConnectionString(mainCtx, "sslmode=disable")
	if err != nil {
		log.Err(err).Msg("creating database")
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Err(err).Msg("opens database")
		return
	}

	grp.Go(func() error {
		service.CreateAndStartHttp(mainCtx, vGenerator, db)
		return nil
	})

	grp.Go(func() error {
		return service.FetchNeighborhoods(nChan)
	})

	grp.Go(func() error {
		for n := range nChan {
			if err := service.GenerateAndWriteVector(mainCtx, vGenerator, db, n); err != nil {
				return fmt.Errorf("generating vector data to database: %w", err)
			}
		}

		return nil
	})

	errChan := make(chan error, 1)

	go func() {
		if err := grp.Wait(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-mainCtx.Done():
		log.Info().Msg("Context cancelled, shutting down...")
	case err := <-errChan:
		log.Err(err).Msg("system has error")
	}

	stopContext, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	if err := pgContainer.Stop(stopContext, new(1*time.Minute)); err != nil {
		log.Err(err).Msg("failed to terminate container")
	}

	close(nChan)
	cancel()
}
