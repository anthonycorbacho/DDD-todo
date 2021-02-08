package adapter

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/anthonycorbacho/DDD-todo/packages/todo/domain"
	"github.com/jmoiron/sqlx"
)

// TodoRepository represents the repository of todos that will manage
// todos in a Database.
type TodoRepository struct {
	db *sqlx.DB
}

// todo represents the internal Database model of a Todo
type todo struct {
	ID      int       `db:"id"`
	Title   string    `db:"title"`
	DueDate time.Time `db:"due_date"`
}

func (t *todo) MapToDomain() domain.Todo {
	return domain.Todo{
		ID:      strconv.Itoa(t.ID),
		Title:   t.Title,
		DueDate: t.DueDate.String(),
	}
}

// NewTodoRepository returns a new todo repository configured to work with the given db.
func NewTodoRepository(db *sqlx.DB) (*TodoRepository, error) {
	if db == nil {
		return nil, errors.New("invalid database")
	}

	return &TodoRepository{
		db: db,
	}, nil
}

func (tr *TodoRepository) Get(ctx context.Context, id int) (domain.Todo, error) {
	if id <= 0 {
		return domain.Todo{}, fmt.Errorf("invalid todo id")
	}

	var t todo
	const q = `SELECT id, title, due_date FROM todo where id = ?`
	if err := tr.db.GetContext(ctx, &t, q, id); err != nil {
		// we might want to do a better error handeling
		// eg: we could return a better error message when the dodo is not found.
		return domain.Todo{}, err
	}

	return t.MapToDomain(), nil
}

func (tr *TodoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	if todo == nil {
		return errors.New("cannot create null todo")
	}

	if len(todo.Title) == 0 {
		return errors.New("cannot create todo without title")
	}

	if len(todo.DueDate) == 0 {
		return errors.New("cannot create todo without due date")
	}

	t, err := time.Parse(time.RFC3339, todo.DueDate)
	if err != nil {
		return err
	}

	const q = `INSERT INTO todo (title, due_date) VALUES (?, ?)`
	res, err := tr.db.ExecContext(ctx, q, todo.Title, t)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	todo.ID = strconv.FormatInt(id, 10)
	return nil
}
