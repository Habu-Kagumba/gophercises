package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"gophercises/url_shortener/urlshort"

	"github.com/boltdb/bolt"
)

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func main() {
	var flagReadFromFormat string

	flag.StringVar(&flagReadFromFormat, "redirect-rules", "db", "Where to read redirect rules from (json / yaml / db).")

	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := make(map[string]string)
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	switch flagReadFromFormat {
	case "json":
		jsonHandler, err := urlshort.JSONHandler(mapHandler)
		if err != nil {
			panic(err)
		}

		fmt.Println("Starting the server on :8080")
		log.Fatal(http.ListenAndServe(":8080", jsonHandler))
	case "yaml":
		yamlHandler, err := urlshort.YAMLHandler(mapHandler)
		if err != nil {
			panic(err)
		}

		fmt.Println("Starting the server on :8080")
		log.Fatal(http.ListenAndServe(":8080", yamlHandler))
	case "db":
		db, err := bolt.Open("./redirect_rules/redirect.db", 0600, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				panic(err)
			}
		}()

		boltHandler, err := urlshort.BoltHandler(db, mapHandler)
		if err != nil {
			panic(err)
		}

		fmt.Println("Starting the server on :8080")
		log.Fatal(http.ListenAndServe(":8080", boltHandler))
	default:
		fmt.Println("Starting the server on :8080")
		log.Fatal(http.ListenAndServe(":8080", mapHandler))
	}
}
