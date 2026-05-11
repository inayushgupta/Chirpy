package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"sort"
	"github.com/google/uuid"
	"github.com/inayushgupta/Chirpy/internal/auth"
	"github.com/inayushgupta/Chirpy/internal/database"
	// "github.com/joho/godotenv"
)

func bodyCleaner(body string) string {

	bad_words := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Split(body, " ")
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
	dbQueries      *database.Queries
	secret         string
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
	Token     string    `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ChipyRed bool `json:"is_chirpy_red"`
	
}

func (cfg *apiConfig) createUsers(w http.ResponseWriter, r *http.Request) {

	// 5 lines to get EMAIL!
	type UserJson struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var user UserJson
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&user)
	if err != nil {
		panic("")
	}

	password_hash, err := auth.HashPassword(user.Password)
	if err != nil {
		panic("")
	}
	// creating new user in the db
	params := database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: password_hash,
	}
	createdUser, err := cfg.dbQueries.CreateUser(r.Context(), params)

	// craete a response
	newUser := User{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	data, err := json.Marshal(newUser)
	if err != nil {
		panic("")
	}

	// write response
	w.WriteHeader(201)
	w.Write(data)
}

type chirp struct {
	Body string `json:"body"`
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {

	// create a chirp if it is valid
	var data chirp
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&data)
	if err != nil {
		respondWithJson(w, 400, ErrorResponse{"something went wrong"})
		return
	}

	token, err := auth.GetBearerToken(r.Header)

	// When the token cannot be optained from the Header
	if err != nil {
 		respondWithJson(w, http.StatusUnauthorized, ErrorResponse{"Unauthorized"})
    	return
	}

	// checking token validity
	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
 		respondWithJson(w, http.StatusUnauthorized, ErrorResponse{"Unauthorized"})
    	return
	}

	if data.Body == "" {
		jsonError := ErrorResponse{"Json Error"}
		respondWithJson(w, 400, jsonError)
	} else if len(data.Body) > 140 {
		respondWithJson(w, 400, ErrorResponse{"Too long!"})
	} else {

		// here the chirp becomes valid

		cleanedBody := bodyCleaner(data.Body)

		params := database.CreateChirpParams{
			Body:   cleanedBody,
			UserID: userId,
		}

		_, err := cfg.dbQueries.GetUser(r.Context(), userId)

		if err != nil {
		}

		createdChirp, err := cfg.dbQueries.CreateChirp(r.Context(), params)

		if err != nil {
			respondWithJson(w, 500, ErrorResponse{"Failed to create chirp: " + err.Error()})
			return
		}

		cleanedRes := SuccessfulRespnse{
			createdChirp.ID,
			createdChirp.CreatedAt,
			createdChirp.UpdatedAt,
			cleanedBody,
			createdChirp.UserID,
		}

		respondWithJson(w, 201, cleanedRes)
	}
}

func (cfg *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	
	authorID := r.URL.Query().Get("author_id")
	sortType := r.URL.Query().Get("sort")

	if sortType != "desc" {
		sortType = "asc"
	}
	
	chirps, err := cfg.dbQueries.GetAllChirps(r.Context())

	if err != nil {
		panic("")
	}
	response := []SuccessfulRespnse{}

	for _, chirp := range chirps {
		the_chirp := SuccessfulRespnse{
			chirp.ID,
			chirp.CreatedAt,
			chirp.UpdatedAt,
			chirp.Body,
			chirp.UserID,
		}

		if authorID == "" || authorID == chirp.UserID.String() {
			response = append(response, the_chirp)
		}
	}

	if sortType == "desc" {
		sort.Slice(response, func(i, j int) bool {
			return response[i].Created_at.After(response[j].Created_at)
		})
	} else {
		sort.Slice(response, func(i, j int) bool {
			return response[i].Created_at.Before(response[j].Created_at) 
		})
	}

	jsoned, err := json.Marshal(response)
	w.Write(jsoned)
}

func (cfg *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("id")
	if chirpId == "" {
		panic("no chirpid suppied")
	}

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
	if err != nil {
		panic("")
	}

	w.Write(jsoned)
}

type userLoginBody struct {
	Password      string `json:"password"`
	Email         string `json:"email"`
}

func (cfg *apiConfig) userLogin(w http.ResponseWriter, r *http.Request) {

	var data userLoginBody
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		panic("")
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), data.Email)
	if err != nil {
		// user does not exist
		w.WriteHeader(401)
		return
	}

	is_same, err := auth.CheckPasswordHash(data.Password, user.HashedPassword)
	if err != nil || !is_same {
		w.WriteHeader(401)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	refreshToken := auth.MakeRefreshToken()

	prm := database.StoreRefreshTokenParams {
		Token: refreshToken,
		CreatedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),

	}
	_ , err = cfg.dbQueries.StoreRefreshToken(
		r.Context(),
		prm,
	)

	if err != nil {
		w.WriteHeader(401)
		return
	}
	// password is correct -- sending login creds
	resp := User{
		user.ID,
		user.CreatedAt,
		user.UpdatedAt,
		user.Email,
		token,
		refreshToken,
		user.IsChirpyRed,
	}

	marshaled, err := json.Marshal(resp)
	if err != nil {
		panic("json marshaler error")
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(marshaled)
}


func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {

	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {w.WriteHeader(410) ;return}

	got, err := cfg.dbQueries.GetRefreshToken(r.Context(), rt)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	if time.Until(got.ExpiresAt) < 0 || got.RevokedAt.Valid == true {
		w.WriteHeader(401)
		return
	}

	newjwt, err := auth.MakeJWT(got.UserID, cfg.secret, time.Hour)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	respondWithJson(w, http.StatusOK, map[string]string{"token": newjwt})

}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {w.WriteHeader(410) ;return}
	
	params := database.UpdateRefreshTokenParams{
		UpdatedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
		RevokedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
		Token: rt,
	}

	_, err = cfg.dbQueries.UpdateRefreshToken(r.Context(), params)
	if err != nil {
		w.WriteHeader(401)
		return
	}

	w.WriteHeader(204)
}


func (cfg *apiConfig) updateEmailPass(w http.ResponseWriter, r *http.Request) {
	// check access using 
	// if access -- wrong -- 401
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithJson(w, 401, ErrorResponse{"token specified? right?"})
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithJson(w, 401, ErrorResponse{"token error"})
		return		
	}

	// get the payload information
	dec := json.NewDecoder(r.Body)
	type x struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	var data x 

	err = dec.Decode(&data)
	if err != nil {
		respondWithJson(w, 401, ErrorResponse{"error decoding"})
		return
	}

	new_pass, _ := auth.HashPassword(data.Password)
	// userinfo, err := cfg.dbQueries.GetUser(r.Context(), userId)
	// if err != nil {
	// 	respondWithJson(w, 401, ErrorResponse{"error getting user info"})
	// }



	userInfo , err := cfg.dbQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		HashedPassword: new_pass,
		Email: data.Email,
		ID: userId,
	})	

	if err != nil {
		respondWithJson(w, 401, ErrorResponse{"error decoding"})
		return
	}

	type ResponseJson struct {
		Email string `json:"email"`
	}

	res := ResponseJson{
		Email: userInfo.Email,
	}

	jsoned, err := json.Marshal(res)

	w.Write(jsoned)

	// get userid and update the password and email
	// hash the password
	// store in database
	// return 200

}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithJson(w, 401, ErrorResponse{"token specified? right?"})
	}

	userId, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithJson(w, 401, ErrorResponse{"token error"})
		return		
	}
	

	chirpIdString := r.PathValue("chirpID")
	chirpID := uuid.MustParse((chirpIdString))

	chirpinfo, err := cfg.dbQueries.GetChirpById(r.Context(), chirpID)

	if err != nil {
		w.WriteHeader(404)
		return
	}

	if chirpinfo.UserID != userId {
		respondWithJson(w, 403, ErrorResponse{"user not the owner of the chirp"})
		return
	}

	_, err = cfg.dbQueries.DeleteChirpById(r.Context(), chirpID)
	if err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
}


type chirpyRedReqShape struct {
	Event string `json:"event"`
	Data struct {
	UserId string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) chirpyRed(w http.ResponseWriter, r *http.Request) {

	apiHeader, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithJson(w, 401, ErrorResponse{err.Error()})
	}

	apiKey := os.Getenv("POLKA_KEY")
	fmt.Println(apiKey)

	if apiKey != apiHeader {
		respondWithJson(w, 404, ErrorResponse{"API KEY MISMATCH"})
	}

	dec := json.NewDecoder(r.Body)
	var data chirpyRedReqShape
	err = dec.Decode(&data)
	if err != nil {w.WriteHeader(401);return}

	if data.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	userinfo, err := cfg.dbQueries.GetUser(r.Context(), uuid.MustParse(data.Data.UserId))

	if err != nil {
		// user is not found
		w.WriteHeader(404)
		return
	}

	if userinfo.IsChirpyRed == true {
		// user already subscribed
		w.WriteHeader(204)
		return
	}

	_, err = cfg.dbQueries.UpdateRedToUser(r.Context(), uuid.MustParse(data.Data.UserId))
	if err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(204)
}