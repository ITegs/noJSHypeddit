package app

import (
	"fmt"
	"net/http"

	"github.com/ITegs/noJSHypeddit/database"
	"github.com/ITegs/noJSHypeddit/renderer"
	"github.com/julienschmidt/httprouter"
)

type App interface {
	Main()
}

type app struct {
	db       database.DB
	renderer renderer.Renderer
}

func NewApp(DB database.DB, Renderer renderer.Renderer) App {
	app := &app{
		db:       DB,
		renderer: Renderer,
	}

	return app
}

func (app *app) Main() {
	fmt.Println("Program started!")

	api := app.buildApi()

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

func (app *app) buildApi() *httprouter.Router {
	router := httprouter.New()

	var routes = []*Route{
		{
			Method:  http.MethodGet,
			Path:    "/",
			Handler: http.HandlerFunc(app.index),
		},
		{
			Method:  http.MethodGet,
			Path:    "/s/:songId",
			Handler: http.HandlerFunc(app.song),
		},
		{
			Method:  http.MethodGet,
			Path:    "/spotify",
			Handler: http.HandlerFunc(app.spotifyRedirect),
		},
	}

	for i := 0; i < len(routes); i++ {
		r := routes[i]
		router.Handler(r.Method, r.Path, r.Handler)
	}

	return router
}

func (app *app) index(w http.ResponseWriter, r *http.Request) {
	app.renderer.Execute("index", w, nil)
}

type PageData struct {
	Name    string
	Artist  string
	Cover   string
	Spotify string
}

func (app *app) song(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request on song endpoint")
	p := httprouter.ParamsFromContext(r.Context())
	id := p.ByName("songId")
	if id == "" {
		fmt.Println("No id given")
		return
	}
	fmt.Println("Got a valid id: ", id)
	song, err := app.db.GetSong(id)
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

	app.renderer.Execute("song", w, data)
	app.db.AddClick(id, "songId")
}

func (app *app) spotifyRedirect(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request on spotify endpoint")
	id := r.URL.Query().Get("id")
	http.Redirect(w, r, "https://open.spotify.com/intl-de/track/"+id, http.StatusFound)
	app.db.AddClick(id, "spotifyId")
}
