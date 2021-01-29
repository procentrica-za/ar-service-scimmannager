package main

import "github.com/gorilla/mux"

type TokenResponse struct {
	Accesstoken  string `json:"access_token"`
	Refreshtoken string `json:"refresh_token"`
	Scopes       string `json:"scope"`
}

type UserResponse struct {
	Message      string `json:"message"`
	Accesstoken  string `json:"access_token"`
	Refreshtoken string `json:"refresh_token"`
	Scopes       string `json:"scope"`
}

type User struct {
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	Scopes    []Scope `json:"scopes"`
	KeySecret string  `json:"keysecret"`
}

type Scope struct {
	Scope string `json:"scope"`
}

type getUserIDResponse struct {
	Resources []IdentityServerResponse `json:"Resources"`
}

type getGroupIDResponse struct {
	Resources []groupResponse `json:"Resources"`
}

type groupResponse struct {
	DisplayName string `json:"displayName"`
	ID          string `json:"id"`
}

type addToGroupResponse struct {
	Success string `json:"success"`
	Message string `json:"message"`
}

type RegisterUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
}

type RegisterUserResponse struct {
	UserCreated string `json:"usercreated"`
	Username    string `json:"username"`
	UserID      string `json:"id"`
	Message     string `json:"message"`
}

type IdentityServerResponse struct {
	ID       string `json:"id"`
	Username string `json:"userName"`
}

type Server struct {
	router *mux.Router
}

type Config struct {
	ISHost          string
	ISPort          string
	APIMHost        string
	APIMPort        string
	ListenServePort string
	ISUsername      string
	ISPassword      string
}
