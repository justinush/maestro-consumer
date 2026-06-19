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
	"github.com/justinush/maestro-consumer/internal/vendor"
	"github.com/justinush/maestro/pkg/run/postgres"
	"github.com/justinush/maestro/pkg/validate"
	"github.com/justinush/maestro/pkg/workflow"
)

func main() {
	ctx := context.Background()

	dbURL := env("DATABASE_URL", "postgres://maestro:maestro@localhost:5433/maestro_consumer?sslmode=disable")
	workflowsDir := env("WORKFLOWS_DIR", "workflows")
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

	reg, err := workflow.LoadDir(workflowsDir, validate.Options{})
	if err != nil {
		log.Fatalf("load workflows from %q: %v", workflowsDir, err)
	}
	for _, key := range reg.List() {
		fmt.Printf("loaded workflow %q v%q\n", key.ID, key.Version)
	}

	runStore := postgres.NewStore(pool)
	appRepo := applicant.NewPostgres(pool)
	vendorStore := vendor.NewPostgres(pool)
	actionReg := kyc.NewActionRegistry(vendorStore)

	svc := kyc.NewService(reg, runStore, appRepo, vendorStore, actionReg)
	handler := kyc.NewHandler(svc)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("maestro-consumer listening on %s (workflows=%s)", addr, workflowsDir)
	log.Fatal(srv.ListenAndServe())
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
