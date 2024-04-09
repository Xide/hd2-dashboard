package migrate

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Migrate(url string, dir string) error {
  m, err := migrate.New(fmt.Sprintf("file://%s", dir), url)
  if err != nil {
    return err
  }
  defer m.Close()
  err = m.Up()
  if err != nil && err != migrate.ErrNoChange {
    return err
  }
  return nil
}
