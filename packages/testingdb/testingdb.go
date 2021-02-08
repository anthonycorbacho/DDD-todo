package testingdb

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/anthonycorbacho/DDD-todo/packages/database"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type TestingDb struct {
	DB   *sqlx.DB
	stop func()
	URL  string
}

var sequence uint32

func (tdb *TestingDb) Open() error {
	tdb.URL = "root:secret1234@tcp(127.0.0.1:3306)/testing"

	now := time.Now().UTC()
	seq := atomic.AddUint32(&sequence, 1) & 0xFFFF
	bytes := make([]byte, 4)
	rand.Seed(now.UnixNano())
	rand.Read(bytes)
	tdb.URL += "_" + now.Format("150405") + "_" + hex.EncodeToString(bytes) + "_" + fmt.Sprintf("%04x", seq)

	cfg, err := mysql.ParseDSN(tdb.URL)
	if err != nil {
		return err
	}
	dbName := cfg.DBName
	cfg.DBName = ""

	// Connect to the root database
	db, stopFunc, err := database.Open(database.MySqlDriver, cfg.FormatDSN())
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}

	// Delete testing databases older than an hour or two
	rows, err := db.Query("SHOW DATABASES")
	if err == nil {
		thisHour := now.Format("15")
		prevHour := now.Add(time.Duration(-time.Hour)).Format("15")
		for rows.Next() {
			var x string
			err = rows.Scan(&x)
			if err == nil &&
				strings.HasPrefix(x, "testing_") &&
				!strings.HasPrefix(x, "testing_"+thisHour) &&
				!strings.HasPrefix(x, "testing_"+prevHour) {
				_, err = db.Exec("DROP DATABASE IF EXISTS " + x)
			}
		}
		rows.Close()
	}

	// Create the testing databases
	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return err
	}

	_, err = db.Exec("USE " + dbName)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile("../../../scripts/todo.sql")
	if err != nil {
		panic(err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		panic(err)
	}

	tdb.DB = db
	tdb.stop = stopFunc
	return nil
}

// Close cleans-up the testing database.
func (tdb *TestingDb) Close() error {
	defer tdb.stop()

	cfg, err := mysql.ParseDSN(tdb.URL)
	if err != nil {
		return err
	}
	dbName := cfg.DBName
	cfg.DBName = ""

	_, _ = tdb.DB.Exec("DROP DATABASE IF EXISTS " + dbName)
	return nil
}
