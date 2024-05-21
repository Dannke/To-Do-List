package main

import (
	"To-Do-List/handlers"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

var mu sync.Mutex

func main() {
	tmpl := template.Must(template.ParseFiles("templates/login.html", "templates/index.html", "templates/register.html"))

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(ctx)

	db := client.Database("todo_db")
	taskCollection := db.Collection("tasks")
	userCollection := db.Collection("users")

	http.HandleFunc("/login", handlers.Login(tmpl, userCollection, &mu))
	http.HandleFunc("/register", handlers.Register(tmpl, userCollection, &mu))
	http.HandleFunc("/logout", handlers.Logout())

	http.HandleFunc("/", handlers.Index(tmpl, taskCollection, &mu))
	http.HandleFunc("/add", handlers.AddTask(taskCollection, &mu))
	http.HandleFunc("/toggle", handlers.ToggleTask(taskCollection, &mu))
	http.HandleFunc("/delete", handlers.DeleteTask(taskCollection, &mu))
	http.HandleFunc("/edit", handlers.EditTask(taskCollection, &mu))

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
