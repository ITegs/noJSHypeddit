package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/julienschmidt/httprouter"
)

type PageData struct {
	Name    string
	Spotify string
}

func initDB() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := rdb.Ping().Err()
	if err != nil {
		fmt.Println("Connection to Redis couldnt be made.")
		panic(err)
	}

	fmt.Println("Connection to Redis established!")

	insertSpotifyId(rdb, "high-tide", "4PkWff16v14sACvFBrKtI0")

	return rdb
}

type apiServer struct {
	template    *template.Template
	redisClient *redis.Client
}

func main() {
	fmt.Println("Program started!")

	fmt.Println("Initializing the DB")
	rdb := initDB()

	fmt.Println("Opening the template")
	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		fmt.Println("Opening template failed")
		return
	}

	apiServer := &apiServer{
		template:    tmpl,
		redisClient: rdb,
	}

	api := apiServer.buildApi()

	server := http.Server{
		Addr:    ":8000",
		Handler: api,
	}

	fmt.Printf("API server listening on port %d\n", 8000)
	server.ListenAndServe()
}

type Route struct {
	Method  string
	Path    string
	Handler http.Handler
}

func (s *apiServer) buildApi() *httprouter.Router {
	router := httprouter.New()

	var routes = []*Route{
		{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: http.HandlerFunc(s.index),
		},
		{
			Method:  http.MethodGet,
			Path:    "/spotify",
			Handler: http.HandlerFunc(s.spotifyRef),
		},
	}

	for i := 0; i < len(routes); i++ {
		r := routes[i]
		router.Handler(r.Method, r.Path, r.Handler)
	}

	return router
}

func (s *apiServer) index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request on index")
	id := r.URL.Query().Get("id")
	if id == "" {
		fmt.Println("No id given")
		return
	}
	fmt.Println("Got a valid id: ", id)
	spotifyId := getSpotifyId(s.redisClient, id)
	fmt.Println("Spotify id: ", spotifyId)

	data := PageData{
		Spotify: spotifyId,
	}

	s.template.Execute(w, data)
}

func (s *apiServer) spotifyRef(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	http.Redirect(w, r, "https://open.spotify.com/intl-de/track/"+id, http.StatusFound)
}

func insertSpotifyId(rdb *redis.Client, userId string, spotifyId string) {
	err := rdb.Set(userId, spotifyId, 0).Err()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted %s into %s\n", spotifyId, userId)
}

func getSpotifyId(rdb *redis.Client, userId string) string {
	value, err := rdb.Get(userId).Result()
	if err != nil {
		panic(err)
	}
	return value
}
