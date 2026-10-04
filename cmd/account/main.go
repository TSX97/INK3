package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TSX97/INK3/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
	"strconv"
	"unsafe"
)

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

var users []User

func search_by_id(id int) *User {
	for i := 0; i < len(users); i++ {
		if users[i].Id == id {
			return &users[i]
		}
	}
	return nil;
}

func newUser(id int, name string) *User {
	return &User{id, name}
}

// === === === ===REST===API=== === === === \\

// GET /users
func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(users)
}

// GET /users{id}
func getUser(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	user := search_by_id(id)
	if user != nil {
		json.NewEncoder(w).Encode(user)
		return
	}
	http.Error(w, "user not found", http.StatusNotFound)
}

// POST /users
func addUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}
	if search_by_id(user.Id) != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	users = append(users, user)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)

}

// PATCH /users/{id}
func patchUserName(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var patch struct {
		Name *string `json:"name"`
	}
	err = json.NewDecoder(r.Body).Decode(&patch)
	if err != nil {
		http.Error(w, "Invalid json", http.StatusBadRequest)
		return
	}

	if patch.Name == nil {
		http.Error(w, "Invalid name", http.StatusBadRequest)
		return
	}

	user := search_by_id(id)
	if user != nil {
		user.Name = *patch.Name
		json.NewEncoder(w).Encode(user)
		return
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

//DELETE /users/{id}
func deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	user := search_by_id(id)
	if user != nil {
		idx := (uintptr(unsafe.Pointer(user)) - uintptr(unsafe.Pointer(&users[0]))) / unsafe.Sizeof(users[0])
		
		users = append(users[:idx], users[idx+1:]...)
		w.WriteHeader(http.StatusNoContent)	
		return
	}
	http.Error(w, "User not found", http.StatusNotFound)
}


//GET /health
func healthChecker(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "database is unhealth", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

}

func main() {

	pool, err := db.NewPool()
	if err != nil {
		panic(err)
	}

	defer pool.Close()

	err = db.Ping(context.Background(), pool)
	if err != nil {
		panic(err)
	}

	fmt.Println("start serve on localhost:8080")
	http.HandleFunc("GET /users", getUsers)
	http.HandleFunc("GET /users/{id}", getUser)
	http.HandleFunc("POST /users", addUser)
	http.HandleFunc("PATCH /users/{id}", patchUserName)
	http.HandleFunc("DELETE /users/{id}", deleteUser)

	http.HandleFunc("GET /health", healthChecker(pool))

	http.ListenAndServe(":8080", nil)
}
