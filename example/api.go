package example

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// ApiError represents an API error with an associated HTTP status code.
type ApiError struct {
	HTTPStatus int
	Err        error
}

func (ae ApiError) Error() string {
	return ae.Err.Error()
}

// User statuses
const (
	statusUser      = 0
	statusModerator = 10
	statusAdmin     = 20
)

// MyApi represents the main API structure.
type MyApi struct {
	statuses map[string]int
	users    map[string]*User
	nextID   uint64
	mu       *sync.RWMutex
}

// NewMyApi creates and initializes a new MyApi instance.
func NewMyApi() *MyApi {
	return &MyApi{
		statuses: map[string]int{
			"user":      statusUser,
			"moderator": statusModerator,
			"admin":     statusAdmin,
		},
		users: map[string]*User{
			"rvasily": {
				ID:       42,
				Login:    "rvasily",
				FullName: "Vasily Romanov",
				Status:   statusAdmin,
			},
		},
		nextID: 43,
		mu:     &sync.RWMutex{},
	}
}

// ProfileParams represents the parameters for the Profile method.
type ProfileParams struct {
	Login string `apivalidator:"required"`
}

// CreateParams represents the parameters for the Create method.
type CreateParams struct {
	Login  string `apivalidator:"required,min=10"`
	Name   string `apivalidator:"paramname=full_name"`
	Status string `apivalidator:"enum=user|moderator|admin,default=user"`
	Age    int    `apivalidator:"min=0,max=128"`
}

// User represents a user in the system.
type User struct {
	ID       uint64 `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	Status   int    `json:"status"`
}

// NewUser represents a newly created user.
type NewUser struct {
	ID uint64 `json:"id"`
}

// apigen:api {"url": "/user/profile", "auth": false}
func (srv *MyApi) Profile(ctx context.Context, in ProfileParams) (*User, error) {
	if in.Login == "bad_user" {
		return nil, fmt.Errorf("bad user")
	}

	srv.mu.RLock()
	user, exist := srv.users[in.Login]
	srv.mu.RUnlock()
	if !exist {
		return nil, ApiError{http.StatusNotFound, fmt.Errorf("user not exist")}
	}

	return user, nil
}

// apigen:api {"url": "/user/create", "auth": true, "method": "POST", "auth_env_key": "MY_API_KEY"}
func (srv *MyApi) Create(ctx context.Context, in CreateParams) (*NewUser, error) {
	if in.Login == "bad_username" {
		return nil, fmt.Errorf("bad user")
	}

	srv.mu.Lock()
	defer srv.mu.Unlock()

	_, exist := srv.users[in.Login]
	if exist {
		return nil, ApiError{http.StatusConflict, fmt.Errorf("user %s exist", in.Login)}
	}

	id := srv.nextID
	srv.nextID++
	srv.users[in.Login] = &User{
		ID:       id,
		Login:    in.Login,
		FullName: in.Name,
		Status:   srv.statuses[in.Status],
	}

	return &NewUser{id}, nil
}

// OtherApi represents another API structure for demonstration purposes.
type OtherApi struct{}

// NewOtherApi creates a new OtherApi instance.
func NewOtherApi() *OtherApi {
	return &OtherApi{}
}

// OtherCreateParams represents the parameters for the OtherApi's Create method.
type OtherCreateParams struct {
	Username string `apivalidator:"required,min=3"`
	Name     string `apivalidator:"paramname=account_name"`
	Class    string `apivalidator:"enum=warrior|sorcerer|rouge,default=warrior"`
	Level    int    `apivalidator:"min=1,max=50"`
}

// OtherUser represents a user in the OtherApi system.
type OtherUser struct {
	ID       uint64 `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
	Level    int    `json:"level"`
}

// apigen:api {"url": "/user/create", "auth": true, "method": "POST", "auth_env_key": "OTHER_API_KEY"}
func (srv *OtherApi) Create(ctx context.Context, in OtherCreateParams) (*OtherUser, error) {
	return &OtherUser{
		ID:       12,
		Login:    in.Username,
		FullName: in.Name,
		Level:    in.Level,
	}, nil
}
