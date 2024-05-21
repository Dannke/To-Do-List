package handlers

import (
	"To-Do-List/models"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"html/template"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func getUserID(r *http.Request) (primitive.ObjectID, error) {
	cookie, err := r.Cookie("user_id")
	if err != nil {
		return primitive.NilObjectID, err
	}

	userID, err := primitive.ObjectIDFromHex(cookie.Value)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return userID, nil
}

func Index(tmpl *template.Template, collection *mongo.Collection, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mu.Lock()
		cursor, err := collection.Find(ctx, bson.M{"userId": userID})
		mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var tasks []models.Task
		if err = cursor.All(ctx, &tasks); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "index.html", tasks)
	}
}

func AddTask(collection *mongo.Collection, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			userID, err := getUserID(r)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			title := r.FormValue("title")
			description := r.FormValue("description")

			if title == "" {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			newTask := models.Task{
				ID:          primitive.NewObjectID(),
				UserID:      userID,
				Title:       title,
				Description: description,
				Status:      false,
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			mu.Lock()
			_, err = collection.InsertOne(ctx, newTask)
			mu.Unlock()

			if err != nil {
				http.Redirect(w, r, "/", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusMethodNotAllowed)
		}
	}
}

func ToggleTask(collection *mongo.Collection, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		idStr := r.URL.Query().Get("id")
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mu.Lock()
		var task models.Task
		err = collection.FindOne(ctx, bson.M{"_id": id, "userId": userID}).Decode(&task)
		if err != nil {
			mu.Unlock()
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		newStatus := !task.Status
		_, err = collection.UpdateOne(ctx, bson.M{"_id": id, "userId": userID}, bson.M{"$set": bson.M{"status": newStatus}})
		mu.Unlock()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func DeleteTask(collection *mongo.Collection, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		idStr := r.URL.Query().Get("id")
		id, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mu.Lock()
		_, err = collection.DeleteOne(ctx, bson.M{"_id": id, "userId": userID})
		mu.Unlock()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func EditTask(collection *mongo.Collection, mu *sync.Mutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			userID, err := getUserID(r)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			idStr := r.FormValue("id")
			id, err := primitive.ObjectIDFromHex(idStr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			title := r.FormValue("title")
			description := r.FormValue("description")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			filter := bson.M{"_id": id, "userId": userID}
			update := bson.M{"$set": bson.M{"title": title, "description": description}}
			mu.Lock()
			_, err = collection.UpdateOne(ctx, filter, update)
			mu.Unlock()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	}
}
