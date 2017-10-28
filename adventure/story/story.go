package story

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

// Story a list of Chapters
type Story map[string]Chapter

// Chapter contains the story
type Chapter struct {
	Title      string   `json:"title"`
	Paragraphs []string `json:"story"`
	Options    []Option `json:"options"`
}

// Option object for every Chapter
type Option struct {
	Text    string `json:"text"`
	Chapter string `json:"arc"`
}

// JSONStory reads a file (json) and returns a Story object
func JSONStory(r io.Reader) (Story, error) {
	d := json.NewDecoder(r)
	var story Story
	if err := d.Decode(&story); err != nil {
		return nil, err
	}

	return story, nil
}

type handler struct {
	once     sync.Once
	filename string
	templ    *template.Template
	s        Story
}

// TemplateHandler parses the templates
func TemplateHandler(s Story) http.Handler {
	return &handler{filename: "index.html", s: s}
}

func defaultPath(r *http.Request) string {
	path := strings.TrimSpace(r.URL.Path)
	if path == "" || path == "/" {
		path = "/intro"
	}
	return path[1:]
}

func (t *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})

	path := defaultPath(r)

	if chapter, ok := t.s[path]; ok {
		if err := t.templ.Execute(w, chapter); err != nil {
			log.Fatal("Failed to parse template:", err)
			http.Error(w, "That wasn't supposed to happen. We have a no-refund policy so ;)", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Chapter not found.", http.StatusNotFound)
}
