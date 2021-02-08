// +build integration

package database

import (
	"context"
	"testing"
)

func TestOpenAndStatusCheck(t *testing.T) {
	// Given
	connection := "todouser:secret1234@tcp(127.0.0.1:3306)/todo"

	db, close, err := Open(MySqlDriver, connection)
	if err != nil {
		t.Errorf("fail opening connection to db: %v", err)
		return
	}
	defer close()

	err = StatusCheck(context.TODO(), db)
	if err != nil {
		t.Errorf("fail statusCkeck: %v", err)
		return
	}
}
