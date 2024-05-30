package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PageData struct {
	Name    string
	Spotify string
}

func initDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017/")
	mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	err = mongoClient.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connection to MongoDB established!")

	if false {
		insertSpotifyId(mongoClient, "high-tide", "4PkWff16v14sACvFBrKtI0")
	}

	return mongoClient, nil
}

type apiServer struct {
	template    *template.Template
	mongoClient *mongo.Client
}

func main() {
	fmt.Println("Program started!")

	fmt.Println("Initializing the DB")
	mCl, err := initDB()
	if err != nil {
		fmt.Println("Connection to MongoDB failed")
		return
	}

	fmt.Println("Opening the template")
	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		fmt.Println("Opening template failed")
		return
	}

	apiServer := &apiServer{
		template:    tmpl,
		mongoClient: mCl,
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
	spotifyId, err := getSpotifyId(s.mongoClient, id)
	if err != nil {
		fmt.Println("No SpotifyId found")
		fmt.Println(err)
		return
	}
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

type Link struct {
	SongID    string `bson:"songId"`
	SpotifyID string `bson:"spotifyId"`
}

func insertSpotifyId(mCl *mongo.Client, song string, spotify string) {
	collection := mCl.Database("noJSHypeddit").Collection("links")
	newLink := Link{
		SongID:    song,
		SpotifyID: spotify,
	}
	result, err := collection.InsertOne(context.TODO(), newLink)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted Song with ID: %s. MongoID: %s\n", song, result.InsertedID)
}

func getSpotifyId(mCl *mongo.Client, song string) (string, error) {
	collection := mCl.Database("noJSHypeddit").Collection("links")
	filter := bson.D{{Key: "songId", Value: song}}
	var result Link
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return "", err
	}

	return result.SpotifyID, nil
}
