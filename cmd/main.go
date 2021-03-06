package main

import (
	"github.com/ful09003/td/components"
	"sync"
	"math/rand"
	_ "strings"
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
	"database/sql"
	"os"
	"log"
	"time"
)

const keySpace = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
var store components.TDStoreInterface

func addTD(w http.ResponseWriter, r *http.Request) {
	urlToSet := r.FormValue("url")
	k := RandStringBytesRmndr(6)
	k1, e := store.Set(components.TDItem{
		Key: k,
		URL: urlToSet,
	})
	if e != nil {
		if e == sql.ErrNoRows {
			//Dup entry
			log.Println("Dupliate request for URL: ", urlToSet)
			w.WriteHeader(302)
			fmt.Fprintln(w, k1)
		} else {
			log.Println(e)
			w.WriteHeader(500)
		}
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintln(w, k1)
	}
}

func redirTD(w http.ResponseWriter, r *http.Request) {
	itemKey := mux.Vars(r)["key"]
	i, e := store.Get(itemKey)

	if e == nil {
		http.Redirect(w, r, i.URL,303)
		return
	}

	switch e {
	case sql.ErrNoRows:
		//Doesn't exist
		w.WriteHeader(404)
		return
	default:
		//Something else
		fmt.Println(e)
		w.WriteHeader(500)
		return
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	envs := getEnv()
	port := envs["port"]

	switch envs["back"] {
	case "mem":
		fmt.Println("Using in-memory store")
		store = components.GenerateNewRAMTDS()
	case "postgres":
		fmt.Println("Using Postgres store")
		connstring := envs["dsn"]
		store = components.GenerateNewPGTDS(connstring)
	default:
		fmt.Println("Defaulting to in-memory store")
		store = components.GenerateNewRAMTDS()
	}

	defer store.Close()

	mux := mux.NewRouter()
	mux.HandleFunc("/n", addTD).Methods("POST")
	mux.HandleFunc("/g/{key}", redirTD).Methods("GET")

	http.ListenAndServe(port, mux)
}


var mutex sync.Mutex

func int63() int64 {
	mutex.Lock()
	v := rand.Int63()
	mutex.Unlock()
	return v
}

func RandStringBytesRmndr(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = keySpace[int63() % int64(len(keySpace))]
	}
	return string(b)
}

func getEnv() map[string]string {
	var r = map[string]string {
		"dsn": "",
		"port": "",
		"back": "",
	}

	if r["back"] = os.Getenv("BACKING_STORE"); r["back"] == "" {
		fmt.Println("No BACKING_STORE env found. Defaulting to mem")
		r["back"] = "mem"
	}

	if r["back"] == "postgres" {
		if r["dsn"] = os.Getenv("DSN"); r["dsn"] == "" {
			panic("DSN required to use a postgres backend")
		}
	}

	if r["port"] = os.Getenv("PORT"); r["port"] == "" {
		r["port"] = ":8080"
	}

	return r
}