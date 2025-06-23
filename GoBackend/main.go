package main

import (
	"GoBackend/config"
	"GoBackend/middleware"
	"log"
	"net/http"
	"os"

	"GoBackend/handlers"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	http.HandleFunc("/check", check)
	http.HandleFunc("/register", handlers.Register)
	http.HandleFunc("/login", handlers.Login)
	http.HandleFunc("/task/create", middleware.AuthMiddleware(handlers.CreateTask))
	http.HandleFunc("/task/list", middleware.AuthMiddleware(handlers.ListAllTask))
	http.HandleFunc("/task/update", middleware.AuthMiddleware(handlers.UpdateTask))
	http.HandleFunc("/task/delete", middleware.AuthMiddleware(handlers.DeletedTask))


	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}
	config.ConnectDB()
	log.Printf("server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func check(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy"}`))
}
