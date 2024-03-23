package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	router.HandleFunc("/account/{id}", makeHTTPHandler(s.handleAccountById))
	router.HandleFunc("/transfer", makeHTTPHandler(s.handleTransfer))
	log.Println("Listening on address", s.listenAddress)

	// Starting the HTTP server with the provided address and router.
	http.ListenAndServe(s.listenAddress, router)
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := &TransferRequest{}
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transferReq)
}

// handleAccount handles requests for account operations.
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("unsupported method: %s", r.Method)
}

func (s *APIServer) handleAccountById(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccountById(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	if r.Method == "PATCH" {
		return s.handleUpdateAccount(w, r)
	}

	return fmt.Errorf("unsupported method: %s", r.Method)
}

// handleGetAccount handles GET requests for retrieving an account.
func (s *APIServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)

	if err != nil {
		return err
	}

	account, err := s.store.GetAccountById(id)

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account) // Writing a JSON response with a dummy Account.
}

// handleGetAccounts handles GET requests for retrieving all accounts.
func (s *APIServer) handleGetAccount(w http.ResponseWriter) error {
	accounts, err := s.store.GetAccounts()

	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
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
	id, err := getId(r)

	if err != nil {
		return err
	}

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, nil)
}

func (s *APIServer) handleUpdateAccount(w http.ResponseWriter, r *http.Request) error {
	updateAccountRequest := &UpdateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(updateAccountRequest); err != nil {
		return err
	}
	id, err := getId(r)
	if err != nil {
		return err
	}
	if err := s.store.UpdateAccount(id, updateAccountRequest); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, nil)
}

// apiFunc is a function signature for API handlers.
type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
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

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid account ID: %s", idStr)
	}
	return id, nil
}
