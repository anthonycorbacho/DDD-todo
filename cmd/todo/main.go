package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/jaeger"
	"github.com/anthonycorbacho/DDD-todo/packages/database"
	"github.com/anthonycorbacho/DDD-todo/packages/todo"
	"github.com/anthonycorbacho/DDD-todo/packages/todo/adapter"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"go.opencensus.io/trace"
)

// main: everything start from here
func main() {
	// Define Log
	log := log.New(os.Stdout, "TODO : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// Configuration of Todo application
	var cfg struct {
		Addr   string `default:"localhost:8080"`
		Jaeger struct {
			Collector string `default:"http://127.0.0.1:14268/api/traces"`
			Agent     string `default:"localhost:6831"`
		}
		SQL struct {
			Driver string `default:"mysql"`
			Addr   string `default:"todouser:secret1234@tcp(127.0.0.1:3306)/todo?parseTime=true"`
		}
	}

	if err := envconfig.Process("todo", &cfg); err != nil {
		log.Fatalf("process config: %v", err)
	}

	// Start tracing
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: cfg.Jaeger.Collector,
		AgentEndpoint:     cfg.Jaeger.Agent,
		Process: jaeger.Process{
			ServiceName: "todo",
			Tags:        []jaeger.Tag{},
		},
	})
	if err != nil {
		panic(err)
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// Start Database
	db, closeFun, err := database.Open(cfg.SQL.Driver, cfg.SQL.Addr)
	if err != nil {
		log.Fatalf("open database connection: %v", err)
	}
	defer closeFun()

	todoRepository, err := adapter.NewTodoRepository(db)
	if err != nil {
		log.Fatalf("creating todo repository: %v", err)
	}

	r := mux.NewRouter().StrictSlash(true)
	todoService, err := todo.New(todoRepository)
	if err != nil {
		log.Fatalf("creating todo service: %v", err)
	}
	// Init routes for TODO service
	todoService.InitRoutes(r)

	// HTTP server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	serverError := make(chan error, 1)
	go func() {
		serverError <- server.ListenAndServe()
	}()

	select {
	case err := <-serverError:
		log.Fatalf("service stopped: %v", err)
	case sig := <-shutdown:
		log.Printf("shutting down the service: '%v'", sig)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("shutting down fail due to error '%v'", err)
		}
	}
}
