package todo

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/anthonycorbacho/DDD-todo/packages/todo/domain"
	"github.com/gorilla/mux"
	"go.opencensus.io/trace"
)

//TodoService represents the service that provides
// the interactions with the todos.
type Service struct {
	repository domain.TodoRepository
}

// New creates a new Todo Service.
func New(repository domain.TodoRepository) (*Service, error) {
	if repository == nil {
		return nil, errors.New("invalid todo repository")
	}

	return &Service{
		repository: repository,
	}, nil
}

func (s *Service) InitRoutes(r *mux.Router) {
	subrouter := r.PathPrefix("/todos").Subrouter()
	subrouter.HandleFunc("", s.Create).Methods("POST")
	subrouter.HandleFunc("/{id}", s.Get).Methods("GET")
}

// Get gets a TODO by its ID.
func (s *Service) Get(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "todo.Get")
	defer span.End()

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid todo id")
		return
	}

	// call the storage to get the todo
	todo, err := s.repository.Get(ctx, id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "todo not found")
		return
	}
	respondWithJSON(w, http.StatusOK, todo)
}

// Create creates a new todo
func (s *Service) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := trace.StartSpan(r.Context(), "todo.Create")
	defer span.End()

	var todo domain.Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid todo format")
		return
	}

	if err := s.repository.Create(ctx, &todo); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, todo)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
