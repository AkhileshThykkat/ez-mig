package runner

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/AkhileshThykkat/ez-mig/internal/session"
	"github.com/golang-migrate/migrate/v4"

	// Register drivers with blank imports so they compile into the binary
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Action string

const (
	ActionUp      Action = "up"
	ActionDown    Action = "down"
	ActionGoto    Action = "goto"
	ActionForce   Action = "force"
	ActionVersion Action = "version"
)

func Execute(s session.Session, action Action, arg *int) error {

	parsedURL, err := url.Parse(s.DbURI)
	if err != nil {
		return fmt.Errorf("invalid database URI: %w", err)
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "postgres" && scheme != "postgresql" && scheme != "mysql" {
		return fmt.Errorf("unsupported database scheme %q: ez-mig v1 only supports postgres and mysql", scheme)
	}

	sourceURL := s.MigrationFilesPath
	if !strings.HasPrefix(sourceURL, "file://") {
		sourceURL = "file://" + sourceURL
	}

	m, err := migrate.New(sourceURL, s.DbURI)
	if err != nil {
		return fmt.Errorf("failed to initialize migration engine: %w", err)
	}

	defer m.Close()

	switch action {
	case ActionUp:
		if arg != nil {
			err = m.Steps(*arg)
		} else {
			err = m.Up()
		}

	case ActionDown:
		if arg != nil {
			err = m.Steps(-*arg)
		} else {
			err = m.Steps(-1)
		}

	case ActionGoto:
		if arg == nil {
			return errors.New("goto command requires a version argument")
		}
		err = m.Migrate(uint(*arg))

	case ActionForce:
		if arg == nil {
			return errors.New("force command requires a version argument")
		}
		err = m.Force(*arg)

	case ActionVersion:
		v, dirty, vErr := m.Version()
		if vErr != nil {
			if errors.Is(vErr, migrate.ErrNilVersion) {
				fmt.Println("No migrations have been applied yet.")
				return nil
			}
			return fmt.Errorf("failed to fetch migration version: %w", vErr)
		}
		dirtyStr := ""
		if dirty {
			dirtyStr = " (DIRTY)"
		}
		fmt.Printf("Current Version: %d%s\n", v, dirtyStr)
		return nil

	default:
		return fmt.Errorf("unhandled migration action: %s", action)
	}

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("Database is already up to date.")
			return nil
		}
		return err
	}

	fmt.Printf("Migration %s executed successfully.\n", action)
	return nil
}
