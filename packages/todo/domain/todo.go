package domain

import "context"

// Todo represents a user todo list
type Todo struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	DueDate string `json:"due_date"`
}

// TodoRepository represents the storage unit of a todo
type TodoRepository interface {
	Get(ctx context.Context, id int) (Todo, error)
	Create(ctx context.Context, todo *Todo) error
}
