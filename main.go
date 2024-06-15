package main

import (
	"github.com/ITegs/noJSHypeddit/app"
	"github.com/ITegs/noJSHypeddit/database"
	"github.com/ITegs/noJSHypeddit/renderer"
)

func main() {
	collection := database.CollectionFactory("noJSHypeddit", "links")
	db := database.NewDB(collection)

	pages := renderer.PagesFactory("./static/index.html", "./static/song.html")
	renderer := renderer.NewRenderer(pages)
	app := app.NewApp(db, renderer)

	app.Main()
}
