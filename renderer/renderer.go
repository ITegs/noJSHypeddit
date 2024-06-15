package renderer

import (
	"html/template"
	"io"
)

type pages struct {
	index *template.Template
	song  *template.Template
}

type renderer struct {
	pages *pages
}

type pageType = string

const (
	indexPage pageType = "index"
	songPage  pageType = "song"
)

type Renderer interface {
	Execute(pageType pageType, w io.Writer, data any)
}

func NewRenderer(pages *pages) Renderer {
	renderer := &renderer{
		pages: pages,
	}

	return renderer
}

func PagesFactory(indexPath string, songPath string) *pages {
	index, err := template.ParseFiles(indexPath)
	if err != nil {
		return nil
	}
	song, err := template.ParseFiles(songPath)
	if err != nil {
		return nil
	}

	return &pages{
		index: index,
		song:  song,
	}
}

func (r *renderer) Execute(pageType pageType, w io.Writer, data any) {
	if pageType == indexPage {
		r.pages.index.Execute(w, data)
	} else if pageType == songPage {
		r.pages.song.Execute(w, data)
	}
}
