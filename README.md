## Description

Simple TODO project following DDD architecture

### Run Application

```
make
./dist/todo-darwin
```

You can now cURL
 - Get a todo: `GET localhost:8080/todos/{id}`
 - Create a todo: `POST localhost:8080/todos` with the following payload
    ```
   {
        "title": "fff",
        "due_date": "2021-02-08T22:04:05Z"
    }
    ```

You can see the tracing by call Jaeger endpoint: `http://localhost:16686`

### Run test

```
make test
```

### Run integration test

Run `docker-compose up`
then
```
make integration
```