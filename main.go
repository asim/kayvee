package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/asim/kv/server"
)

//go:embed html/*
var html embed.FS

var (
	nodes   = flag.String("nodes", "", "comma seperated list of nodes")
	address = flag.String("address", ":4001", "http host:port")

	// global server
	srv *server.Server
)

func init() {
	flag.Parse()
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")
	val := r.Form.Get("val")

	if err := srv.Set(key, val); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	if err := srv.Delete(key); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.Form.Get("key")

	val, err := srv.Get(key)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte(val.(string)))
}

func main() {
	var members []string
	if len(*nodes) > 0 {
		members = strings.Split(*nodes, ",")
	}

	// create new server
	s, err := server.New(&server.Options{
		Members: members,
	})

	if err != nil {
		log.Fatal(err)
	}

	// set global server
	srv = s

	log.Printf("Local node %s\n", srv.Address())

	// set http handlers
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/del", delHandler)
	http.HandleFunc("/get", getHandler)

	// extract the embedded html directory
	htmlContent, err := fs.Sub(html, "html")
	if err != nil {
		log.Fatal(err)
	}

	// serve the html directory by default
	http.Handle("/", http.FileServer(http.FS(htmlContent)))

	log.Printf("Listening on %s\n", *address)

	if err := http.ListenAndServe(*address, nil); err != nil {
		log.Fatal(err)
	}
}
