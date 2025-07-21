package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Todo represents a todo item
type Todo struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Completed   bool               `json:"completed" bson:"completed"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// TodoHandler handles todo-related HTTP requests
type TodoHandler struct {
	collection *mongo.Collection
}

// NewTodoHandler creates a new TodoHandler
func NewTodoHandler(collection *mongo.Collection) *TodoHandler {
	return &TodoHandler{
		collection: collection,
	}
}

// CreateTodo handles POST /todos
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var todo Todo
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if todo.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Check if title already exists
	var existingTodo Todo
	err := h.collection.FindOne(context.Background(), bson.M{"title": todo.Title}).Decode(&existingTodo)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Todo with this title already exists",
			"code":  "DUPLICATE_TITLE",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to check title uniqueness",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	// Set timestamps
	todo.CreatedAt = time.Now()
	todo.UpdatedAt = time.Now()

	// Insert into MongoDB
	result, err := h.collection.InsertOne(context.Background(), todo)
	if err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// Set the ID from the insert result
	todo.ID = result.InsertedID.(primitive.ObjectID)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

// GetTodos handles GET /todos
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cursor, err := h.collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var todos []Todo
	if err := cursor.All(context.Background(), &todos); err != nil {
		http.Error(w, "Failed to decode todos", http.StatusInternalServerError)
		return
	}

	// Return empty array if no todos found
	if todos == nil {
		todos = []Todo{}
	}

	json.NewEncoder(w).Encode(todos)
}

// GetTodo handles GET /todos/{id}
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid todo ID",
			"code":  "INVALID_ID",
		})
		return
	}

	var todo Todo
	err = h.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&todo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Todo not found",
				"code":  "NOT_FOUND",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to fetch todo",
				"code":  "DATABASE_ERROR",
			})
		}
		return
	}

	json.NewEncoder(w).Encode(todo)
}

// UpdateTodo handles PUT /todos/{id}
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid todo ID",
			"code":  "INVALID_ID",
		})
		return
	}

	var updateData Todo
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON",
			"code":  "INVALID_JSON",
		})
		return
	}

	// Validate required fields
	if updateData.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Title is required",
			"code":  "MISSING_TITLE",
		})
		return
	}

	// Check if title already exists (excluding current todo)
	var existingTodo Todo
	err = h.collection.FindOne(context.Background(), bson.M{
		"title": updateData.Title,
		"_id":   bson.M{"$ne": id},
	}).Decode(&existingTodo)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Todo with this title already exists",
			"code":  "DUPLICATE_TITLE",
		})
		return
	} else if err != mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to check title uniqueness",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	// Set updated timestamp
	updateData.UpdatedAt = time.Now()

	// Create update document
	update := bson.M{
		"$set": bson.M{
			"title":       updateData.Title,
			"description": updateData.Description,
			"completed":   updateData.Completed,
			"updated_at":  updateData.UpdatedAt,
		},
	}

	// Update the document
	result, err := h.collection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to update todo",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	if result.MatchedCount == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Todo not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	// Fetch and return the updated todo
	var updatedTodo Todo
	err = h.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&updatedTodo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Failed to fetch updated todo",
			"code":  "DATABASE_ERROR",
		})
		return
	}

	json.NewEncoder(w).Encode(updatedTodo)
}

func (h *TodoHandler) UpdateTodoStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	var statusUpdate struct {
		Completed bool `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&statusUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"completed":  statusUpdate.Completed,
			"updated_at": time.Now(),
		},
	}

	result, err := h.collection.UpdateOne(context.Background(), bson.M{"_id": id}, update)
	if err != nil {
		http.Error(w, "Failed to update todo status", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	var updatedTodo Todo
	err = h.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&updatedTodo)
	if err != nil {
		http.Error(w, "Failed to fetch updated todo", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedTodo)
}

// DeleteTodo handles DELETE /todos/{id}
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	result, err := h.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// connectMongoDB establishes connection to MongoDB
func connectMongoDB() (*mongo.Client, error) {
	// Get MongoDB URI from environment variable, fallback to localhost
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

// createUniqueIndex creates a unique index on the title field
func createUniqueIndex(collection *mongo.Collection) error {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"title", 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return err
	}

	fmt.Println("Created unique index on title field")
	return nil
}

func main() {
	// Connect to MongoDB
	client, err := connectMongoDB()
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.Background())

	// Get collection
	collection := client.Database("todoapp").Collection("todos")

	// Create unique index on title field
	if err := createUniqueIndex(collection); err != nil {
		log.Printf("Warning: Failed to create unique index: %v", err)
	}

	// Create handler
	todoHandler := NewTodoHandler(collection)

	// Setup routes
	r := mux.NewRouter()

	// Add CORS middleware first
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	api := r.PathPrefix("/api/v1").Subrouter()

	// Todo routes
	api.HandleFunc("/todos", todoHandler.CreateTodo).Methods("POST")
	api.HandleFunc("/todos", todoHandler.GetTodos).Methods("GET")
	api.HandleFunc("/todos/{id}", todoHandler.GetTodo).Methods("GET")
	api.HandleFunc("/todos/{id}", todoHandler.UpdateTodo).Methods("PUT")
	api.HandleFunc("/todos/{id}/status", todoHandler.UpdateTodoStatus).Methods("PATCH")
	api.HandleFunc("/todos/{id}", todoHandler.DeleteTodo).Methods("DELETE")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(addr, r))
}
