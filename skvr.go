package main

import (
	"log"
	"net/http"
	"os"
)

func getENV(k string) string {
	def := ""
	v, ok := os.LookupEnv(k)
	if !ok {
		def = " (default value)"
		switch k {
		case "SKVR_DIR":
			{
				v = "/var/lib/skvr"
			}
		case "SKVR_DEFAULT_NAMESPACE":
			{
				v = "default"
			}
		case "SKVR_INDEX_KEY":
			{
				v = "index.html"
			}
		case "SKVR_PORT":
			{
				v = "8077"
			}
		}
	}
	log.Printf("Config: %s = %s%s", k, v, def)
	return v
}

func main() {
	defer db.Close()
	for {
		err := server()
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func server() error {
	port := getENV("SKVR_PORT")
	mux := http.NewServeMux()
	mux.HandleFunc("/", api)
	log.Printf("Simple KV Rest Server starting on port %s...", port)
	return http.ListenAndServe("0.0.0.0:"+port, mux)
}
