package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddress string
	store         Storage // Storage interface for interacting with data store.
}

func NewAPIServer(address string, store Storage) *APIServer {
	return &APIServer{
		listenAddress: address, // Initializing APIServer with provided address and store.
		store:         store,
	}
}

// WriteJSON writes JSON response to the client.
func WriteJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json") // Setting response header to indicate JSON content.
	w.WriteHeader(status)                              // Setting the response status code.
	return json.NewEncoder(w).Encode(v)                // Encoding provided data as JSON and writing to response.
}

func (s *APIServer) Run() {
	router := mux.NewRouter() // Creating a new router instance using gorilla/mux.

	// Registering handlers for specific routes.
	router.HandleFunc("/account", makeHTTPHandler(s.handleAccount))
	router.HandleFunc("/account/{id}", makeHTTPHandler(s.handleGetAccount))

	log.Println("Listening on address", s.listenAddress)

	// Starting the HTTP server with the provided address and router.
	http.ListenAndServe(s.listenAddress, router)
}

// handleAccount handles requests for account operations.
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}
	return fmt.Errorf("unsupported method: %s", r.Method)
}

// handleGetAccount handles GET requests for retrieving an account.
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"] // Extracting the "id" parameter from the request URL.

	fmt.Println(id) // Logging the extracted ID.

	return WriteJSON(w, http.StatusOK, &Account{}) // Writing a JSON response with a dummy Account.
}

// handleCreateAccount handles POST requests for creating an account.
func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountRequest := &CreateAccountRequest{}

	if err := json.NewDecoder(r.Body).Decode(createAccountRequest); err != nil {
		return err
	}

	account := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName)

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

// handleDeleteAccount handles DELETE requests for deleting an account.
func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// apiFunc is a function signature for API handlers.
type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string
}

// makeHTTPHandler wraps an API handler function with error handling.
func makeHTTPHandler(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Invoking the provided handler function and handling any error.
		if err := fn(w, r); err != nil {
			// If an error occurs, writing an error response with HTTP status Bad Request.
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
