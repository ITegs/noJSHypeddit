package main

import (
	"github.com/ITegs/noJSHypeddit/app"
	"github.com/ITegs/noJSHypeddit/database"
)

func main() {
	collection := database.CollectionFactory("noJSHypeddit", "links")
	db := database.NewDB(collection)

	app := app.NewApp(db)

	app.Main()
}
