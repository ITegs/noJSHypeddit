package app

import (
	"fmt"
	"net/http"
	"text/template"

	"github.com/ITegs/noJSHypeddit/database"
	"github.com/julienschmidt/httprouter"
)

type App interface {
	Main()
}

type app struct {
	pages *pages
	db    *database.DB
}

func NewApp(DB *database.DB) App {
	app := &app{
		db: DB,
	}

	return app
}

type pages struct {
	index *template.Template
	song  *template.Template
}

type apiServer struct {
}

func (app *app) Main() {
	fmt.Println("Program started!")

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
