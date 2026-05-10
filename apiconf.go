package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/google/uuid"
	"sync/atomic"
	"strings"
	"encoding/json"
	"github.com/inayushgupta/Chirpy/internal/database"
)

func bodyCleaner(body string) string {

	bad_words := map[string]bool{
        "kerfuffle": true,
        "sharbert":  true,
        "fornax":    true,
    }

	words := strings.Split(body , " ")
	new_body := ""
	
	for _, word := range words {
		if _, ok := bad_words[strings.ToLower(word)]; ok {
			word = "****"
		}
		new_body = new_body + " " + word
	}

	return new_body[1:]

}


type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries 
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) getCount(w http.ResponseWriter, r *http.Request) {
	template := 
`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>
`	
	formatted := fmt.Sprintf(template, cfg.fileserverHits.Load())
	w.Write([]byte(formatted))
	w.Header().Set("Content-Type", "text/html")
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	cfg.dbQueries.DeleteAllUsers(r.Context())
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {
	
	// 5 lines to get EMAIL!
	type UserJson struct {Email string `json:"email"`}
	var user UserJson
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&user)
	if err != nil {panic("")}

	// creating new user in the db
	createdUser, err := cfg.dbQueries.CreateUser(r.Context(), user.Email)

	// craete a response
	newUser := User {
		ID: createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email: createdUser.Email,
	}

	data, err := json.Marshal(newUser)
	if err != nil {panic("")}

	// write response
	w.WriteHeader(201) 
	w.Write(data)
}

type chirp struct {
	Body string `json:"body"`
	User_id string `json:"user_id"`
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r*http.Request) {

	// create a chirp if it is valid
	var data chirp
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&data)
	
	if err != nil {	
		generalError := ErrorResponse{"Something went wrong"}
		respondWithJson(&w, 400, generalError)
	} else {
		if data.Body == "" {
			jsonError := ErrorResponse{"Json Error"}
			respondWithJson(&w, 400, jsonError)
		} else if len(data.Body) > 140 {
			respondWithJson(&w, 400, ErrorResponse{"Too long!"})
		} else {

			// here the chirp becomes valid

			cleanedBody := bodyCleaner(data.Body)

			userID, parseErr := uuid.Parse(data.User_id)
			if parseErr != nil {
				respondWithJson(&w, 400, ErrorResponse{"Invalid user_id"})
				return
			}

			params := database.CreateChirpParams{
				Body: cleanedBody,
				UserID: userID,
			}

			_, err := cfg.dbQueries.GetUser(r.Context(), userID)

			if err != nil {}

			createdChirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)

			if err != nil {
				respondWithJson(&w, 500, ErrorResponse{"Failed to create chirp: " + err.Error()})
				return
			}



			cleanedRes := SuccessfulRespnse{
				createdChirp.ID,
				createdChirp.CreatedAt,
				createdChirp.UpdatedAt,
				cleanedBody,
				createdChirp.UserID,
			}

			respondWithJson(&w, 201, cleanedRes)
		}
	}
}



func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetAllChirps(r.Context())
	
	if err != nil {panic("")}
	response := []SuccessfulRespnse{}

	for _, chirp := range chirps {
		the_chirp := SuccessfulRespnse{
			chirp.ID,
			chirp.CreatedAt,
			chirp.UpdatedAt,
			chirp.Body,
			chirp.UserID,
		}
		response = append(response, the_chirp)
	}

	jsoned, err := json.Marshal(response)
	w.Write(jsoned)
}

func (cfg *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("id")
	if chirpId == "" { panic("no chirpid suppied") }
	fmt.Println(chirpId)

	chirpRow, err := cfg.dbQueries.GetChirpById(r.Context(), uuid.MustParse(chirpId))

	if err != nil {
		w.WriteHeader(404)
		return 
	}

	responseData := Chirp{
		chirpRow.ID,
		chirpRow.CreatedAt,
		chirpRow.UpdatedAt,
		chirpRow.Body,
		chirpRow.UserID,
		}
	
	jsoned, err := json.Marshal(responseData)
	if err != nil {panic("")}

	w.Write(jsoned)
}