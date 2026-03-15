package main

import (
	"net/http"
	"os"
	"ybg-backend-copy/api" // Import folder api kamu sebagai package
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Karena di api/index.go kamu sudah ada var router yang di-init
	// Kita bisa langsung memanggil Handler-nya
	http.HandleFunc("/", api.Handler)
	http.ListenAndServe(":"+port, nil)
}
