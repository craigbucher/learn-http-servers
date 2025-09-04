package main

import (
	"fmt"
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"github.com/craigbucher/learn-http-servers/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // The underscore tells Go that you're importing it for its side effects, not because you need to use it
)

// Create a struct in main.go that will hold any stateful, in-memory data we'll need to keep track of. 
// In our case, we just need to keep track of the number of requests we've received:
type apiConfig struct {
	// add a single field to the struct named fileserverHits with type atomic.Int32:
	fileserverHits atomic.Int32
	// create a new *database.Queries, and store it in your apiConfig struct so that handlers can access it:
	db *database.Queries
	// 
	platform       string
}

func main() {
	const port = "8080"
	const filepathRoot = "."
	
	//  call godotenv.Load() at the beginning of your main() function to load the .env file into 
	// your environment variables:
	godotenv.Load()
	// use os.Getenv to get the DB_URL from the environment:
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	// 
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	} else {
		fmt.Printf("Platform is %s\n", platform)
	}

	// Next, sql.Open() a connection to your database:
	// "postgres" is the name of the driver to use; available from _ "github.com/lib/pq"
	// dbURL is the Postgres connection string
	dbConn, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	// temp, for testing:
	if err := dbConn.Ping(); err != nil {
    // handle ping/connectivity error
	} else {
		fmt.Printf("Successfully connected to database\n")
	}

	// Use your SQLC generated database package to create a new *database.Queries:
	dbQueries := database.New(dbConn)

	// create a new variable called apiCfg that uses a composite literal to create a new instance of 
	// the apiConfig struct:
	apiCfg := apiConfig{
		// initialize the fileserverHits field with a new atomic.Int32 value
		// atomic.Int32{} creates an atomic.Int32 starting at the default value (zero)
		// This ensures you can immediately use its atomic methods like .Add() and .Load()
		fileserverHits: atomic.Int32{},
		// assigns dbQueries (the database connection) to the db field so handlers can run queries:
		db:             dbQueries,
		platform:       platform,
	}

	// Create a new http.ServeMux:
	mux := http.NewServeMux()
	// Update the fileserver to use the /app/ path instead of /:
	// Not only will you need to mux.Handle the /app/ path, you'll also need to strip the /app prefix 
	// from the request path before passing it to the fileserver handler:
		// * mux.Handle("/app/", ...) registers a handler for any HTTP request that begins with /app/. That means requests to /app/, /app/index.html, /app/assets/logo.png, etc., will all go to the handler you provide.
		// * http.StripPrefix("/app", ...) middleware strips the /app prefix from the URL path before handing it off to the next handler. For example, a request to /app/assets/logo.png would be turned into just /assets/logo.png for the next handler.
		// * http.FileServer(http.Dir(filepathRoot)) serves static files from the directory specified by the variable filepathRoot.
		// * apiCfg.middlewareMetricsInc(...) wraps the file-serving handler with your middlewareMetricsInc middleware, which *increments your hit counter each time the /app/ route is accessed*
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	
	// Add the Readiness Endpoint:
	// (I recommend using the mux.HandleFunc to register your handler.)
		// * HandleFunc is a method on mux that lets you register a new route
		// * "/healthz" is the route (or path) you’re registering. Any HTTP request to the /healthz URL will be handled by what you specify next.
		// * handlerReadiness is the function that will run whenever someone calls /healthz. This function must match the signature func(http.ResponseWriter, *http.Request)
	// Update the following paths to only accept GET requests:
	// prepend /api to the beginning of each of our API endpoints:
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	// // Use the http.NewServeMux's .Handle() method to add a handler for the root path (/):
	// // Use a standard http.FileServer as the handler:
	// // Use http.Dir to convert a filepath (in our case a dot: . which indicates the current directory) 
	// // to a directory for the http.FileServer:
	// mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	mux.HandleFunc("POST /api/users", apiCfg.handlerUsersCreate)
	// Add a POST /api/chirps handler:
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerChirpsRetrieve)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChirpsGet)
	mux.HandleFunc("POST /api/login", apiCfg.handlerLogin)

	// Register the handlerMetrics handler with the serve mux on the /metrics path:
	// Update the following paths to only accept GET requests:
		// prepend /api to the beginning of each of our API endpoints:
	// Swap out the GET /api/metrics endpoint, which just returns plain text, for a GET /admin/metrics 
	// that returns HTML:
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	// create and register a handler on the /reset path that, when hit, will reset your fileserverHits 
	// back to 0:
	// Update the /reset endpoint to only accept POST requests:
		// prepend /api to the beginning of each of our API endpoints:
		// Update the POST /api/reset to POST /admin/reset:
	// Update the POST /admin/reset endpoint to delete all users in the database (but don't mess with the schema)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	// Create a new http.Server struct:
	srv := &http.Server{
		Addr:    ":" + port,	// Set the .Addr field to ":8080"
		Handler: mux,			// Use the new "ServeMux" as the server's handler
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	// Use the server's ListenAndServe method to start the server:
	log.Fatal(srv.ListenAndServe())
}

// Your handler can just be a function that matches the signature of http.HandlerFunc:
	// * w http.ResponseWriter: is an interface provided by Go. It represents the response that will be 
	// sent back to the client. You use w to write headers, status codes, and body content for the HTTP response
	// * r *http.Request: is a pointer to the http.Request struct, which represents all the information 
	// about the incoming HTTP request (like method, headers, body, URL, etc.)
	// (The function signature must match Go’s required handler standard - Even if you don’t need the r 
	// parameter in your code, you must include it to satisfy the interface. )
// func handlerReadiness(w http.ResponseWriter, r *http.Request) {
func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	// Write the Content-Type header:
		// * w.Header returns a map-like object representing the HTTP response headers; you use it to set 
		// or modify headers before sending the response.
	// w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	// use the Content-Type header to set the response type to text/html so that the browser knows 
	// how to render it:
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	// Write the status code using w.WriteHeader:
		// w.WriteHeader Sets the HTTP status code for the response (e.g., 200, 404, 503); You must call it before 
		// writing the response body, or it will default to 200 (OK) if you write data without calling it.
	w.WriteHeader(http.StatusOK)
	// Write the body text using w.Write:
		// * http.StatusOK is a constant with the value 200—the standard HTTP status code for a successful request.
		// * http.StatusText(http.StatusOK) turns that status code into its textual representation: it returns the string "OK".
		// * []byte(...) converts the string "OK" into a byte slice, which is the format required by w.Write.
		// * w.Write(...) writes those bytes to the HTTP response body. So, in effect, you’re sending OK as the body of the HTTP response.
	// w.Write([]byte(http.StatusText(http.StatusOK)))
	// Create a new handler that writes the number of requests that have been counted as plain text to 
	// the HTTP response:
	htmlContent := fmt.Sprintf(`<html>
		<body>
			<h1>Welcome, Chirpy Admin</h1>
			<p>Chirpy has been visited %d times!</p>
		</body>
	</html>`, cfg.fileserverHits.Load())
	//w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
	w.Write([]byte(htmlContent))
	}
	// You typically call w.Header().Set(...) first, then w.WriteHeader(...), then write the body.


	// write a new middleware method on a *apiConfig that increments the fileserverHits counter 
	// every time it's called:
		// This defines a method on the *apiConfig struct
		// It takes another handler (next http.Handler) as an argument
		// It returns a new handler (http.Handler)—specifically, a wrapped version that does extra work
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// This returns a new http.HandlerFunc. That’s essentially an anonymous function that itself 
	// handles HTTP requests:
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// increments the atomic counter by 1 each time the handler is called (which means: every 
		// matching request):
		cfg.fileserverHits.Add(1)
		// call the original (wrapped) handler to actually process the HTTP request:
		next.ServeHTTP(w, r)
	})

}

/////////////////////////////  Would you like a quick quiz on stateful handlers