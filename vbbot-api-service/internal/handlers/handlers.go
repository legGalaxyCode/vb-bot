package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"vbbot-api-service/internal/helpers"
	"vbbot-api-service/pkg/logging"
)

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type RequestPayload struct {
	Action   string          `json:"action"`
	Auth     AuthPayload     `json:"auth,omitempty"`
	Register RegisterPayload `json:"register,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember string `json:"remember,omitempty"`
}

type RegisterPayload struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func PingHandle(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = helpers.WriteJSON(w, http.StatusOK, payload)
}

func HandleSubmission(w http.ResponseWriter, r *http.Request) {
	logger := logging.GetLogger()
	var requestPayload RequestPayload

	logger.Println("read json from responseWriter")
	err := helpers.ReadJSON(w, r, &requestPayload)
	if err != nil {
		helpers.ErrorJSON(w, err)
		logger.Warnf("error in reading json: %v; %s;", err, requestPayload)
		return
	}

	logger.Println("proceed action:", requestPayload.Action)
	switch requestPayload.Action {
	case "auth":
		authenticate(w, requestPayload.Auth)
	case "register":
		register(w, requestPayload.Register)
	default:
		helpers.ErrorJSON(w, errors.New("unknown action"))
	}
}

func authenticate(w http.ResponseWriter, auth AuthPayload) {
	logger := logging.GetLogger()
	logger.Println("start authentication")
	// create json we will send to mservice
	jsonData, _ := json.MarshalIndent(auth, "", "\t")
	// call the service
	request, err := http.NewRequest("POST", "http://localhost:8081/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Printf("error in request: %s\n", err)
		helpers.ErrorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		logger.Printf("error in getting response: %s\n", err)
		helpers.ErrorJSON(w, err)
		return
	}
	defer response.Body.Close()
	// make sure we get back correct status code
	if response.StatusCode == http.StatusUnauthorized {
		logger.Println("invalid credentials, status unauthorized")
		helpers.ErrorJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		logger.Printf("error in calling auth service, response status code:%d\n", response.StatusCode)
		helpers.ErrorJSON(w, errors.New("error calling auth service"))
		return
	}

	// read body response
	var jsonFromService jsonResponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		logger.Printf("error decode json:%s\n", err)
		helpers.ErrorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		helpers.ErrorJSON(w, err, http.StatusUnauthorized)
		logger.Printf("error in jsonfromservice:%s\n", err)
		return
	}

	var payload jsonResponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	helpers.WriteJSON(w, http.StatusAccepted, payload)
	logger.Println("auth end with response:", payload)
}

func register(w http.ResponseWriter, reg RegisterPayload) {
	logger := logging.GetLogger()
	logger.Println("start register")
	payload := jsonResponse{
		Error:   false,
		Message: "Registered",
		Data:    "",
	}
	helpers.WriteJSON(w, http.StatusAccepted, payload)
	logger.Println("finish register")
}
