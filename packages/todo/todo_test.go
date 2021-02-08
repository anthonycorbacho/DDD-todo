package todo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/anthonycorbacho/DDD-todo/packages/todo/domain"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type MockStorage struct {
	mu sync.Mutex
	m  map[string]domain.Todo
}

func (ms *MockStorage) Get(ctx context.Context, id int) (domain.Todo, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	todo, ok := ms.m[strconv.Itoa(id)]
	if !ok {
		return todo, fmt.Errorf("todo not found")
	}
	return todo, nil
}

func (ms *MockStorage) Create(ctx context.Context, todo *domain.Todo) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if len(todo.ID) == 0 {
		// for testing we generate a predictable id
		todo.ID = fmt.Sprintf("%s-%s", todo.Title, todo.DueDate)
	}

	// injected fail case
	if todo.Title == "genError" {
		return errors.New("error happen")
	}

	ms.m[todo.ID] = *todo
	return nil
}

func Test_Get(t *testing.T) {
	t.Parallel()

	mock := MockStorage{
		m: map[string]domain.Todo{},
	}
	duedate := time.Date(2021, time.July, 17, 0, 0, 0, 0, time.UTC)
	// insert data
	mock.m["123"] = domain.Todo{ID: "123", Title: "Hello", DueDate: duedate.String()}

	todoService, _ := New(&mock)

	var cases = []struct {
		name        string
		input       string // id of the todo
		output      domain.Todo
		expectError bool
		httpStatus  int
	}{
		{
			name:        "Todo not found",
			input:       "4583489573489",
			expectError: true,
			httpStatus:  http.StatusNotFound,
		},
		{
			name:  "Todo found",
			input: "123",
			output: domain.Todo{
				ID:      "123",
				Title:   "Hello",
				DueDate: duedate.String(),
			},
			httpStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/todos/"+tc.input, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			r := mux.NewRouter()
			todoService.InitRoutes(r)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.httpStatus, rr.Code)
			if !tc.expectError {
				var output domain.Todo
				_ = json.Unmarshal(rr.Body.Bytes(), &output)
				assert.Equal(t, tc.output, output)
			}
		})
	}
}

func Test_Create(t *testing.T) {
	t.Parallel()
	mock := MockStorage{
		m: map[string]domain.Todo{},
	}
	todoService, _ := New(&mock)

	var cases = []struct {
		name        string
		input       domain.Todo
		output      domain.Todo
		expectError bool
		httpStatus  int
	}{
		{
			name: "create new todo",
			input: domain.Todo{
				Title:   "test1",
				DueDate: "date1",
			},
			output: domain.Todo{
				ID:      "test1-date1",
				Title:   "test1",
				DueDate: "date1",
			},
			httpStatus: http.StatusCreated,
		},
		{
			name: "create new todo fail",
			input: domain.Todo{
				Title:   "genError",
				DueDate: "date1",
			},
			output:      domain.Todo{},
			expectError: true,
			httpStatus:  http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, _ := json.Marshal(tc.input)
			req, err := http.NewRequest("POST", "/todos", bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			r := mux.NewRouter()
			todoService.InitRoutes(r)
			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.httpStatus, rr.Code)
			if !tc.expectError {
				var output domain.Todo
				_ = json.Unmarshal(rr.Body.Bytes(), &output)
				assert.Equal(t, tc.output, output)
			}
		})
	}

}
