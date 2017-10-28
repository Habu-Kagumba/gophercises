package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"gophercises/adventure/story"
)

func main() {
	port := flag.String("port", ":3000", "server port")
	filename := flag.String("story", "story/stories.json", "JSON file with the story.")
	flag.Parse()
	fmt.Printf("Using the story in %s.\n", *filename)

	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal("Failed to open file:", err)
	}

	s, err := story.JSONStory(file)
	if err != nil {
		log.Fatal("Failed to decode json:", err)
	}

	h := story.TemplateHandler(s)
	fmt.Println("Server running at:", *port)
	log.Fatal(http.ListenAndServe(*port, h))
}
