package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessfulRespnse struct {
	Id         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_id    uuid.UUID `json:"user_id"`
}

type Chirp struct {
	Id         uuid.UUID `json:"id"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
	Body       string    `json:"body"`
	User_id    uuid.UUID `json:"user_id"`
}

func respondWithJson(w http.ResponseWriter, statusCode int, respJson any) {
	data, err := json.Marshal(respJson)
	if err != nil {
		panic("")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}
