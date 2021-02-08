package database

import (
	"context"
	"database/sql"
	"errors"
	"net/url"
	"time"

	"contrib.go.opencensus.io/integrations/ocsql"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.opencensus.io/trace"

	// MySQL driver
	_ "github.com/go-sql-driver/mysql"
	// Postgres driver
	_ "github.com/lib/pq"
)

const (
	MySqlDriver    = "mysql"
	PostgresDriver = "postgres"
)

// Open knows how to open a database connection based on the driver and connection string.
func Open(driver, connection string) (*sqlx.DB, func(), error) {
	// verify if the driver is supported and valid.
	switch driver {
	case "mysql":
		cfg, err := mysql.ParseDSN(connection)
		if err != nil {
			return nil, nil, err
		}
		if cfg.Params == nil {
			cfg.Params = map[string]string{}
		}
		cfg.Params["parseTime"] = "true"
		cfg.Params["timeout"] = "4s"
		cfg.Params["readTimeout"] = "8s"
		cfg.Params["writeTimeout"] = "8s"
		connection = cfg.FormatDSN()
		if _, err := mysql.ParseDSN(connection); err != nil {
			return nil, nil, errors.New("invalid mysql connection string")
		}

	case "postgres":
		if _, err := url.Parse(connection); err != nil {
			return nil, nil, errors.New("invalid postgres connection string")
		}

	default:
		return nil, nil, errors.New("unsupported database driver: " + driver)

	}

	ocsql.RegisterAllViews()
	// Register our ocsql wrapper for the provided database driver.
	driverName, err := ocsql.Register(driver, ocsql.WithAllTraceOptions())
	if err != nil {
		return nil, nil, err
	}

	// Connect to th database using the ocsql driver wrapper.
	db, err := sql.Open(driverName, connection)
	if err != nil {
		return nil, nil, err
	}
	// Record DB stats every 5 seconds until we exit.
	stop := ocsql.RecordStats(db, 5*time.Second)
	dbx := sqlx.NewDb(db, driver)

	closefn := func() {
		stop()
		_ = dbx.Close() // explicitly ignore error.
	}

	return dbx, closefn, nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *sqlx.DB) error {
	ctx, span := trace.StartSpan(ctx, "database.StatusCheck")
	defer span.End()

	// Run a simple query to determine connectivity. The db has a "Ping" method
	// but it can false-positive when it was previously able to talk to the
	// database but the database has since gone away. Running this query forces a
	// round trip to the database.
	const q = `SELECT true`
	var tmp bool
	return db.QueryRowContext(ctx, q).Scan(&tmp)
}
