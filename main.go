package main

import _ "github.com/lib/pq"
import (
	"database/sql"
	"fmt"
	"github.com/inayushgupta/Chirpy/internal/database"
	"github.com/joho/godotenv"
	"net/http"
	"os"
)

func main() {
	godotenv.Load()

	var apiconf apiConfig = apiConfig{
		secret: os.Getenv("SECRET"),
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic("")
	}
	apiconf.dbQueries = database.New(db)
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	fun := apiconf.middlewareMetricsInc(http.StripPrefix("/app", fileServer))
	mux.Handle("/app/", fun)

	// admin route
	mux.HandleFunc("GET /admin/metrics", apiconf.getCount)
	mux.HandleFunc("POST /admin/reset", apiconf.reset)
	// api route
	mux.HandleFunc("GET /api/healthz", healthz)
	mux.HandleFunc("POST /api/users", apiconf.createUsers)
	mux.HandleFunc("POST /api/chirps", apiconf.createChirp)
	mux.HandleFunc("GET /api/chirps", apiconf.getAllChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiconf.getChirpById)
	mux.HandleFunc("POST /api/login", apiconf.userLogin)
	mux.HandleFunc("POST /api/refresh", apiconf.refresh)
	mux.HandleFunc("/api/revoke", apiconf.revoke)
	mux.HandleFunc("PUT /api/users",apiconf.updateEmailPass)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiconf.deleteChirp)
	
	// webhooks - they are nothing but apis triggered 
	// 			  by external services (automatically)
	
	mux.HandleFunc("POST /api/polka/webhooks", apiconf.chirpyRed)

	var server http.Server
	server.Handler = mux
	server.Addr = ":8080"

	fmt.Println("Starting Server @8080")
	server.ListenAndServe()
}
