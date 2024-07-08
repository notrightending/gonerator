# Gonerator

Gonerator is a simple Go code generator that creates handlers based on method definitions and comments.
It's a simple project designed to showcase capacity of code generation in Go and also it might save you a bit of time writing boilerplate. 

## Features

- Generates HTTP handlers from Go struct methods
- Supports GET and POST methods
- Implements environment variable-based authentication
- Provides basic request parameter validation

## Usage

1. Clone the repository:
   ```
   git clone https://github.com/notrightending/gonerator.git
   cd gonerator
   ```

2. Build the generator:
   ```
   go build -o gonerator cmd/generator/main.go
   ```


3. Define your API methods in a Go file:

```go
// apigen:api {"url": "/user/create", "auth": true, "method": "POST", "auth_env_key": "MY_API_KEY"}
func (api *MyAPI) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
    // Your implementation here
}
```

4. Run the generator:

```
./gonerator input.go output.go
```

5. Use the generated handlers in your main application.

## Validation Tags

The generator supports the following validation tags:

- `required`: Field must not be empty
- `min`: Minimum value (for int) or length (for string)
- `max`: Maximum value (for int) or length (for string)
- `enum`: List of allowed values
- `default`: Default value if not provided

Example:
```go
type CreateUserParams struct {
    Username string `apivalidator:"required,min=3"`
    Age      int    `apivalidator:"min=18,max=99"`
    Role     string `apivalidator:"enum=user|admin,default=user"`
}
```

## Note

This generator requires the `ApiError` struct to be defined in your project:

```go
type ApiError struct {
    HTTPStatus int
    Err        error
}
```


## Testing

If you're feeling ambitious

```
go test ./test -v
```
