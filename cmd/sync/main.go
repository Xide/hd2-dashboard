package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Xide/helldivers2-dashboard/pkg/client"
	migrate "github.com/Xide/helldivers2-dashboard/pkg/migrations"
	"github.com/doug-martin/goqu/v9"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/time/rate"


	"database/sql"

	_ "github.com/lib/pq"
)


var flags *pflag.FlagSet = pflag.NewFlagSet("hde", pflag.ExitOnError)

func initLogger() {
  logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
  }))
  slog.SetDefault(logger)
}

func init() {
  viper.SetEnvPrefix("hde")
  viper.AutomaticEnv()

  flags.String("api_url", "https://api.live.prod.thehelldiversgame.com/api", "URL of the API")
  flags.String("json_data_dir", "/data", "Directory where the static json data is stored")
  flags.String("postgres_url", "postgres://postgres:password@localhost:5432/helldivers2?sslmode=disable", "URL of the postgres database")
  flags.String("migrations_dir", "/migrations", "Directory where the migration files are stored")
  err := viper.BindPFlags(flags)
  if err != nil {
    panic(err)
  }
  initLogger()
}

type ManagerSharedState struct {
  limiter *rate.Limiter
  client *client.ClientWithResponses
  db *sql.DB
}

type NewsManagerRequest struct {
  WarID int
}

func ReconcileBatch(ctx context.Context, state ManagerSharedState, fromTimestamp int, warID int) (int, int, error) {
  lang := "en-US"
  news, err := state.client.GetNewsFeedWarIdWithResponse(ctx, warID, &client.GetNewsFeedWarIdParams{
    FromTimestamp: &fromTimestamp,
    AcceptLanguage: &lang,
  })
  if err != nil {
    slog.Error("failed to get news", slog.Any("error", err))
    return 0, fromTimestamp, err
  }
  if news.StatusCode() != 200 {
    slog.Error("failed to get news", slog.Any("status_code", news.StatusCode()))
    return 0, fromTimestamp, fmt.Errorf("failed to get news: %d", news.StatusCode())
  }
  if len(*news.JSON200) > 0 {
    slog.Debug("Got news", slog.String("manager", "news"), slog.Int("war_id", warID), slog.Int("count", len(*news.JSON200)))
    stmt := goqu.Insert("news").Cols("war_id", "id", "published_at", "message")
    stmtVals := [][]interface{}{}
    for _, n := range *news.JSON200 {
      stmtVals = append(stmtVals, goqu.Vals{
        warID,
        n.Id,
        n.Published,
        n.Message,
      })
    }
    sql, _, err := stmt.Vals(stmtVals...).OnConflict(goqu.DoNothing()).ToSQL()
    if err != nil {
      slog.Error("failed to build sql", slog.Any("error", err))
      return 0, fromTimestamp, err
    }
    _, err = state.db.Exec(sql)
    if err != nil {
      slog.Error("failed to insert news", slog.Any("error", err))
      return 0, fromTimestamp, err
    }
    fromTimestamp = int((*news.JSON200)[len(*news.JSON200)-1].Published)
  } else {
    slog.Debug("No news", slog.String("manager", "news"), slog.Int("war_id", warID))
  }
  return len(*news.JSON200), fromTimestamp, nil
}

func NewsManagerReconcile(ctx context.Context, state ManagerSharedState, req NewsManagerRequest) error {
  slog.Debug("Reconciling", slog.String("manager", "news"), slog.Any("war_id", req.WarID))
  tStart := time.Now()
  state.limiter.Wait(ctx)
  tEnd := time.Now()
  slog.Debug("Rate limiter wait time", slog.String("manager", "news"), slog.Int("war_id", req.WarID), slog.Any("wait_time", tEnd.Sub(tStart).Seconds()))
  fromTimestamp := 0
  err := state.db.QueryRow("SELECT COALESCE(MAX(published_at), 0) FROM news WHERE war_id = $1", req.WarID).Scan(&fromTimestamp)
  if err != nil {
    slog.Error("failed to get max published_at", slog.Any("error", err))
    return err
  }
  slog.Debug("Got max published_at", slog.String("manager", "news"), slog.Int("war_id", req.WarID), slog.Int("from_timestamp", fromTimestamp))
  reconcilied, newFromTimestamp, err := ReconcileBatch(ctx, state, fromTimestamp + 1, req.WarID)
  if err != nil {
    slog.Error("failed to reconcile batch", slog.Any("error", err))
    return err
  }
  slog.Info("Reconciled", slog.String("manager", "news"), slog.Any("war_id", req.WarID), slog.Int("reconciled", reconcilied), slog.Int("from_timestamp", newFromTimestamp))
  return nil
}

func main() {
  slog.Info("Performing database migrations", slog.String("migration_dir", viper.GetString("migrations_dir")))
  err := migrate.Migrate(viper.GetString("postgres_url"), viper.GetString("migrations_dir"))
  if err != nil {
    slog.Error("failed to run database migrations", slog.Any("error", err))
    os.Exit(1)
  }
  slog.Info("Database migrations complete")
  client, err := client.NewClientWithResponses(viper.GetString("api_url"))
  if err != nil {
    slog.Error("failed to create client", slog.Any("error", err))
    os.Exit(1)
  }

  db, err := sql.Open("postgres", viper.GetString("postgres_url"))
  if err != nil {
    slog.Error("failed to open database", slog.Any("error", err))
    os.Exit(1)
  }
  defer db.Close()

  limiter := rate.NewLimiter(1, 5)
  slog.Info("Starting news manager")

  for {
    err = NewsManagerReconcile(context.Background(), ManagerSharedState{
      limiter: limiter,
      client: client,
      db: db,
    }, NewsManagerRequest{
      WarID: 801,
    })
    if err != nil {
      slog.Error("failed to reconcile news manager", slog.Any("error", err))
      os.Exit(1)
    }
    time.Sleep(60 * time.Second)
  }
}

