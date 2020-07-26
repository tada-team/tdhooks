package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"gopkg.in/yaml.v2"
)

func main() {
	configPathPtr := flag.String("config", "/etc/tdhooks/default.yml", "path to config")
	flag.Parse()

	b, err := ioutil.ReadFile(*configPathPtr)
	if err != nil {
		log.Println("config error:", err)
		os.Exit(1)
	}

	var config struct {
		Listen  string   `yaml:"listen"`
		Servers []server `yaml:"servers"`
		Poker   pokerConfig
	}

	if err := yaml.Unmarshal(b, &config); err != nil {
		log.Println("config error:", err)
		os.Exit(1)
	}

	rtr := mux.NewRouter()
	config.Poker.listen(rtr, config.Servers)

	srv := http.NewServeMux()
	srv.Handle("/", rtr)

	server := &http.Server{
		Addr:    config.Listen,
		Handler: srv,
	}

	if server.Addr == "" {
		server.Addr = "127.0.0.1:8042"
	}

	log.Printf("start tdhooks at: http://%s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Println("start server fail:", err)
		os.Exit(1)
	}
}
