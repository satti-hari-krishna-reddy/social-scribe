package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"social-scribe/backend/internal/models"
	repo "social-scribe/backend/internal/repositories"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func SignupUserHandler(resp http.ResponseWriter, req *http.Request) {
	if req.Body == nil {
		http.Error(resp, `{"error": "Failed to parse credintials: body is empty"}`, http.StatusBadRequest)
		return
	}
	user := models.User{}

	err := json.NewDecoder(req.Body).Decode(&user)
	if err != nil {
		http.Error(resp, `{"error": "Bad request: unable to decode JSON"}`, http.StatusBadRequest)
		return
	}

	user.UserName = strings.TrimSpace(user.UserName)
	user.UserName = strings.Join(strings.Fields(strings.ToLower(user.UserName)), "")
	user.PassWord = strings.TrimSpace(user.PassWord)

	if len(user.UserName) < 4 || len(user.UserName) > 64{
		http.Error(resp, `{"error": "The username should contain a minimum of 4 and maximum of 64 characters"}`, http.StatusBadRequest)
		return
	}
	if len(user.PassWord) < 8 || len(user.PassWord) > 128{
		http.Error(resp, `{"error": "The password should contain a minimum of 8 and maximun of 128 characters"}`, http.StatusBadRequest)
		return
	}

	existingUser, err := repo.GetUserByName(user.UserName)
	if err != nil {
		http.Error(resp, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		log.Printf("[ERROR] Error checking existing user: %v", err)
		return
	}
	if existingUser != nil {
		http.Error(resp, `{"message" : "Username already taken"}`, http.StatusConflict)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PassWord), bcrypt.DefaultCost)
	if err != nil {
		http.Error(resp, `{"error": "Internal server error"}`, http.StatusInternalServerError)
		log.Printf("[ERROR] Error hashing password for user '%s': %v", user.UserName, err)
		return
	}
	user.PassWord = string(hashedPassword)
	userId, err := repo.InsertUser(user)
	if err != nil {
		log.Printf("[ERROR] Unable to create user %v: %v", user.UserName, err)
		http.Error(resp, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	log.Printf("[INFO] User '%s' successfully registered with ID: %s", user.UserName, userId)

	resp.WriteHeader(http.StatusCreated)
	resp.Write([]byte(`{"message" : "User successfully registered"}`))
}

func LoginUserHandler(resp http.ResponseWriter, req *http.Request) {

	if req.Body == nil {
		http.Error(resp, `{"error": "Failed to parse login credentials: body is empty"}`, http.StatusBadRequest)
		return
	}
	data := models.LoginStruct{}

	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		http.Error(resp, `{"error": "Bad request: unable to decode JSON"}`, http.StatusBadRequest)
		return
	}

	data.Username = strings.ToLower(strings.TrimSpace(data.Username))
	if len(data.Username) < 4 || len(data.Username) > 64 {
		http.Error(resp, `{"error": "Username is should in range of minimum 4 to maximum 64 characters}`, http.StatusBadGateway)
	}
	if len(data.Password) > 128 {
		http.Error(resp, `{"error" : "password is too long, the maximum allowed length is 128 chars"}`, http.StatusBadGateway)
	}
	user, err := repo.GetUserByName(data.Username)
	if user == nil {
		http.Error(resp, `{"success": false, "reason": "Username and/or password is incorrect"}`, http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("[ERROR] Failed to get user for the username %s and the error is %s", data.Username, err)
		http.Error(resp, `{"error" : "Internal server error"}`, http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PassWord), []byte(data.Password))
	if err != nil {
		http.Error(resp, `{"success": false, "reason": "Username and/or password is incorrect"}`, http.StatusBadRequest)
		return
	}

    resp.WriteHeader(http.StatusAccepted)
	resp.Write([]byte(`{"message" : "user successfully authenticated"}`))
}

func GetUserNotificationsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0{
		http.Error(resp, `{"error": "cant able parse id field, reason is missing id field in the request"}`, http.StatusBadRequest)
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to find user for the id: %s and error is %s", userId, err)
		http.Error(resp,`{"error": ""}`,http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(resp,`{"error": "user id is not valid"}`, http.StatusNotFound)
		return
	}
	respone := map[string]interface{}{
		"notifications" : user.Notifications,
	}
	responseJson, err := json.Marshal(respone)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success": false, "reason": "Failed unpacking"}`))
		return
	}

	resp.WriteHeader(200)
	resp.Write(responseJson)

}
 func GetUserSharedBlogsHandler(resp http.ResponseWriter, req *http.Request){
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success" : false, "reason" : "user id not found in the request}`))
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] Failed to find user for the id: %s and error is %s", userId, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : false}`))
		return
	}
	if user == nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"success" : false, "reason" : "user id is invalid"}`))
		return
	}
	response := map[string]interface{}{
		"shared_blogs" : user.SharedBlogs,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`{"sucess" : false, "reason" : "Failed unpacking}`))
		return
	}
	resp.WriteHeader(200)
	resp.Write(responseJson)

}

 func GetUserScheduledBlogsHandler(resp http.ResponseWriter, req *http.Request){
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`"success" : false, "reason" : "user id not provided`))
		return

	}
	user, err := repo.GetUserById(userId)
	if user == nil {
		resp.WriteHeader(401)
	    resp.Write([]byte(`"success" : false, "reason" : "user id is invalid`))
		return
	}
	if err != nil {
		log.Printf("[ERROR] Failed to find user for the id: %s and error is %s", userId, err)
		resp.WriteHeader(500)
        resp.Write([]byte(`{"success" : "false"}`))
		return
	}
	response := map[string]interface{}{
		"scheduled_blogs" : user.ScheduledBlogs,
	}
	responseJson, err := json.Marshal(response)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : false}`))
		return
	}
	resp.WriteHeader(200)
	resp.Write(responseJson)
}

func ClearUserNotificationsHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	userId := vars["id"]
	if len(userId) == 0 {
		resp.WriteHeader(401)
		resp.Write([]byte(`"success" : false, "reason" : "missing user id in the request"`))
		return
	}
	user, err := repo.GetUserById(userId)
	if err != nil {
		log.Printf("[ERROR] failed to get user for the id: %s and the error is %s", userId, err)
		resp.WriteHeader(500)
		resp.Write([]byte(`{"success" : false}`))
		return
	}
	if user == nil {
		resp.WriteHeader(401)
		resp.Write([]byte(`"success" : false, "reason" : "invalid user id"`))
		return

	}
	user.Notifications = []string{}
	err = repo.UpdateUser(userId, user)
	if err != nil {
		log.Printf("[ERROR] failed to update user with id: %s", userId)
		resp.WriteHeader(500)
	    resp.Write([]byte(`{"success" : false, "reason" : "}`))
		return
	}
	resp.WriteHeader(200)
	resp.Write([]byte(`{"success" : true, "message" : "notifications cleared sucessfully"}`))
}