package urlshort

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v2"
)

var redirectRulesDir = "./redirect_rules/"

func readRedirectRulesFiles() ([]byte, []byte) {
	redirectRuleFiles, err := ioutil.ReadDir(redirectRulesDir)
	if err != nil {
		log.Fatal(err)
	}

	var yamlContents, jsonContents []byte

	for _, file := range redirectRuleFiles {
		switch fileExt := filepath.Ext(file.Name()); fileExt {
		case ".json":
			jsonContents, err = ioutil.ReadFile(redirectRulesDir + file.Name())
			if err != nil {
				log.Fatal(err)
			}
		case ".yml":
			yamlContents, err = ioutil.ReadFile(redirectRulesDir + file.Name())
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return yamlContents, jsonContents
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if dest, ok := pathsToUrls[path]; ok {
			http.Redirect(w, r, dest, http.StatusFound)
			return
		}

		fallback.ServeHTTP(w, r)
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(fallback http.Handler) (http.HandlerFunc, error) {
	var yamlPathURLs []struct {
		Path string `yaml:"path"`
		URL  string `yaml:"url"`
	}

	yamlBytes, _ := readRedirectRulesFiles()
	if err := yaml.Unmarshal(yamlBytes, &yamlPathURLs); err != nil {
		return nil, err
	}

	pathsToUrls := make(map[string]string)
	for _, p := range yamlPathURLs {
		pathsToUrls[p.Path] = p.URL
	}

	return MapHandler(pathsToUrls, fallback), nil
}

// JSONHandler will parse JSON and quack like YAMLHandler
func JSONHandler(fallback http.Handler) (http.HandlerFunc, error) {
	var jsonPathURLs []struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}

	_, jsonBytes := readRedirectRulesFiles()
	if err := json.Unmarshal(jsonBytes, &jsonPathURLs); err != nil {
		return nil, err
	}

	pathsToUrls := make(map[string]string)
	for _, p := range jsonPathURLs {
		pathsToUrls[p.Path] = p.URL
	}

	return MapHandler(pathsToUrls, fallback), nil
}

// BoltHandler read redirect rules from BoltDB and quack like JSONHandler
func BoltHandler(db *bolt.DB, fallback http.Handler) (http.HandlerFunc, error) {
	pathsToUrls := make(map[string]string)
	seedBoltHandler(db)

	if err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("RedirectRules"))
		b.ForEach(func(k, v []byte) error {
			pathsToUrls[string(k[:])] = string(v[:])
			return nil
		})

		return nil
	}); err != nil {
		return nil, err
	}

	return MapHandler(pathsToUrls, fallback), nil
}

// SeedBoltHandler seed the "RedirectRules" Bucket with some test data
func seedBoltHandler(db *bolt.DB) {
	if err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("RedirectRules"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return b.Put([]byte("/github"), []byte("https://github.com/explore"))
	}); err != nil {
		panic(err)
	}
}
