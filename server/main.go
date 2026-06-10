package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	app := NewApp()
	app.Start()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("raaz server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, corsMiddleware(app)))
}
