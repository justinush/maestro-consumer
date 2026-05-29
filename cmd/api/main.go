package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinush/maestro-consumer/internal/applicant"
	"github.com/justinush/maestro-consumer/internal/kyc"
	"github.com/justinush/maestro-consumer/internal/migrate"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run/postgres"
)

func main() {
	ctx := context.Background()

	dbURL := env("DATABASE_URL", "postgres://maestro:maestro@localhost:5433/maestro_consumer?sslmode=disable")
	workflowPath := env("WORKFLOW_PATH", "workflow/kyc.yaml")
	addr := env("ADDR", ":8080")

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := migrate.Up(ctx, pool, "migrations"); err != nil {
		log.Fatal(err)
	}
	if err := postgres.ApplySchema(ctx, pool); err != nil {
		log.Fatal(err)
	}

	rt, err := maestro.Load(workflowPath)
	if err != nil {
		log.Fatal(err)
	}
	if def := rt.Definition(); def != nil {
		fmt.Printf("workflow %q v%q\n", def.ID, def.Version)
	}

	runStore := postgres.NewStore(pool)
	appRepo := applicant.NewPostgres(pool)
	svc := kyc.NewService(rt, runStore, appRepo)
	handler := kyc.NewHandler(svc)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("maestro-consumer listening on %s", addr)
	log.Fatal(srv.ListenAndServe())
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
