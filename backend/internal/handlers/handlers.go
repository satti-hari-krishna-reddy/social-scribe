package handlers

import (
	// "encoding/json"
	"net/http"
)

func GetUserHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte(`{"message" : "hello ra"}`))
}

func RegisterUserHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte (`{"message" : "user sucessfully registered"}`))
}

func UserLoginHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte(`{"message" : "user successfully authenticated"}`))
}