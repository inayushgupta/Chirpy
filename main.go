package main
import _ "github.com/lib/pq"
import (
	"fmt"
	"os"
	"net/http"
	"database/sql"
	"github.com/joho/godotenv"
	"github.com/inayushgupta/Chirpy/internal/database"
)

var apiconf apiConfig

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil { panic("") }
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
	mux.HandleFunc("/api/chirps/{id}", apiconf.getChirpById)

	var server http.Server
	server.Handler = mux
	server.Addr = ":8080"

	fmt.Println("Starting Server @8080")
	server.ListenAndServe()
}
