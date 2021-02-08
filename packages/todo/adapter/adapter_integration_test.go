// +build integration

package adapter

import (
	"context"
	"strconv"
	"testing"

	"github.com/anthonycorbacho/DDD-todo/packages/testingdb"
	"github.com/anthonycorbacho/DDD-todo/packages/todo/domain"
	"github.com/stretchr/testify/assert"
)

func setup() (*testingdb.TestingDb, error) {
	var tdb testingdb.TestingDb
	if err := tdb.Open(); err != nil {
		return nil, err
	}
	return &tdb, nil
}

func TestTodoAdapter(t *testing.T) {
	tdb, err := setup()
	if err != nil {
		assert.Fail(t, err.Error())
	}
	defer tdb.Close()

	repository, err := NewTodoRepository(tdb.DB)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// Create a TODO
	todo := domain.Todo{
		Title:   "my todo 1",
		DueDate: "2021-02-08T22:04:05Z",
	}

	ctx := context.Background()
	err = repository.Create(ctx, &todo)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.NotEmpty(t, todo.ID)

	// Get a TODO
	id, err := strconv.Atoi(todo.ID)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	todo2, err := repository.Get(ctx, id)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, todo.ID, todo2.ID)
	assert.Equal(t, todo.Title, todo2.Title)

}
