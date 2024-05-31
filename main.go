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

type Song struct {
	SongId            string `bson:"songId"`
	Name              string `bson:"name"`
	Artist            string `bson:"artist"`
	Cover             string `bson:"cover"`
	SpotifyId         string `bson:"spotifyId"`
	NumClicksSongId   int    `bson:"numClicks-songId"`
	NumClickSpotifyId int    `bson:"numClicks-spotifyId"`
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

	// hometown := &Song{
	// 	Name:      "Hometown",
	// 	SongID:    "hometown",
	// 	SpotifyID: "47x1Gh7yk5mblUWxWRdtjH",
	// }
	// insertSong(mongoClient, hometown)

	return mongoClient, nil
}

type pages struct {
	index *template.Template
	song  *template.Template
}

type apiServer struct {
	pages       *pages
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

	fmt.Println("Opening the templates")
	index, err := template.ParseFiles("./static/index.html")
	if err != nil {
		fmt.Println("Opening song template failed")
		return
	}
	song, err := template.ParseFiles("./static/song.html")
	if err != nil {
		fmt.Println("Opening song template failed")
		return
	}

	apiServer := &apiServer{
		pages: &pages{
			index: index,
			song:  song,
		},
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
			Path:    "/s/:songId",
			Handler: http.HandlerFunc(s.song),
		},
		{
			Method:  http.MethodGet,
			Path:    "/spotify",
			Handler: http.HandlerFunc(s.spotifyRedirect),
		},
	}

	for i := 0; i < len(routes); i++ {
		r := routes[i]
		router.Handler(r.Method, r.Path, r.Handler)
	}

	return router
}

func (s *apiServer) index(w http.ResponseWriter, r *http.Request) {
	s.pages.index.Execute(w, nil)
}

type PageData struct {
	Name    string
	Artist  string
	Cover   string
	Spotify string
}

func (s *apiServer) song(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request on song endpoint")
	p := httprouter.ParamsFromContext(r.Context())
	id := p.ByName("songId")
	if id == "" {
		fmt.Println("No id given")
		return
	}
	fmt.Println("Got a valid id: ", id)
	song, err := getSong(s.mongoClient, id)
	if err != nil {
		fmt.Println("No Song found")
		fmt.Println(err)
		return
	}

	data := PageData{
		Name:    song.Name,
		Artist:  song.Artist,
		Cover:   song.Cover,
		Spotify: song.SpotifyId,
	}

	s.pages.song.Execute(w, data)
	addClick(s.mongoClient, id, "songId")
}

func (s *apiServer) spotifyRedirect(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request on spotify endpoint")
	id := r.URL.Query().Get("id")
	http.Redirect(w, r, "https://open.spotify.com/intl-de/track/"+id, http.StatusFound)
	addClick(s.mongoClient, id, "spotifyId")
}

func insertSong(mCl *mongo.Client, song *Song) {
	collection := mCl.Database("noJSHypeddit").Collection("links")
	song.NumClickSpotifyId = 0
	result, err := collection.InsertOne(context.TODO(), song)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inserted %s with ID: %s. MongoID: %s\n", song.Name, song.SongId, result.InsertedID)
}

func getSong(mCl *mongo.Client, song string) (*Song, error) {
	collection := mCl.Database("noJSHypeddit").Collection("links")
	filter := bson.D{{Key: "songId", Value: song}}
	var result *Song
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func addClick(mCl *mongo.Client, id string, target string) {
	collection := mCl.Database("noJSHypeddit").Collection("links")
	filter := bson.D{{Key: target, Value: id}}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "numClicks-" + target, Value: 1}}}}
	_, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		fmt.Printf("Register click failed. Target: %s\n", target)
	}
}
