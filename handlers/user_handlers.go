package handlers

import (
	"To-Do-List/models"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"net/http"
	"regexp"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Register(tmpl *template.Template, collection *mongo.Collection, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == "" || password == "" {
				tmpl.ExecuteTemplate(w, "register.html", "All fields are required")
				return
			}

			passwordHash := sha256.New()
			passwordHash.Write([]byte(password))
			hashedPassword := hex.EncodeToString(passwordHash.Sum(nil))

			newUser := models.User{
				Username: username,
				Password: hashedPassword,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			mu.Lock()
			_, err := collection.InsertOne(ctx, newUser)
			mu.Unlock()

			if err != nil {
				tmpl.ExecuteTemplate(w, "register.html", "Registration failed")
				return
			}

			http.Redirect(w, r, "/login", http.StatusSeeOther)
		} else {
			tmpl.ExecuteTemplate(w, "register.html", nil)
		}
	}
}

func Login(tmpl *template.Template, collection *mongo.Collection, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			if username == "" || password == "" {
				tmpl.ExecuteTemplate(w, "login.html", "All fields are required")
				return
			}

			passwordHash := sha256.New()
			passwordHash.Write([]byte(password))
			hashedPassword := hex.EncodeToString(passwordHash.Sum(nil))

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			var user models.User
			mu.Lock()
			err := collection.FindOne(ctx, bson.M{"username": username, "password": hashedPassword}).Decode(&user)
			mu.Unlock()

			if err != nil {
				tmpl.ExecuteTemplate(w, "login.html", "Invalid username or password")
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     "user_id",
				Value:    user.ID.Hex(),
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			tmpl.ExecuteTemplate(w, "login.html", nil)
		}
	}
}

func Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:    "user_id",
			Value:   "",
			Path:    "/",
			Expires: time.Now().Add(-1 * time.Hour),
		})
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func isValidInput(input string) bool {
	validInput := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validInput.MatchString(input)
}
