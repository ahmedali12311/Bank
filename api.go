package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gorilla/mux"
)

type apiServer struct {
	listenaddr string
	store      storage
}

func NewApiServer(listServ string, store storage) *apiServer {
	return &apiServer{
		listenaddr: listServ,
		store:      store,
	}
}

func (s *apiServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/login", httpHandler(s.handleLogin))

	router.HandleFunc("/account", httpHandler(s.UserHandler))
	router.HandleFunc("/account/{id}", withJWTAuth(httpHandler(s.GetAccountbyID), s.store))
	router.HandleFunc("/transfer", httpHandler(s.TransferTo))

	http.ListenAndServe(s.listenaddr, router)
}

func (s *apiServer) handleLogin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not allowed %s", r.Method)
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	acc, err := s.store.GetAccountbyNumber(int(req.Number))
	if err != nil {
		return err
	}

	if !acc.ValidPassword(req.Password) {
		return fmt.Errorf("not authenticated")
	}
	token, err := createJWT(acc)
	if err != nil {
		return err
	}
	resp := loginResponse{
		Token:  token,
		Number: int64(acc.Number),
	}
	return WriteJSON(w, http.StatusOK, resp)
}

func (s *apiServer) UserHandler(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetUser(w, r)
	}
	if r.Method == "POST" {
		return s.CreateUser(w, r)
	}

	if r.Method == "DELETE" {
		return s.DeleteUser(w, r)
	}
	return fmt.Errorf("Method Not Allowed %s", r.Method)
}

func (s *apiServer) GetUser(w http.ResponseWriter, r *http.Request) error {
	account, _ := s.store.GetAllAccounts()

	return WriteJSON(w, http.StatusOK, account)
}
func (s *apiServer) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var Request = CreateAccountRequest{}
	if err := json.NewDecoder(r.Body).Decode(&Request); err != nil {
		return err
	}
	Account, err := NewAccount(Request.FirstName, Request.LastName, Request.Password)
	if err != nil {
		return err
	}
	if err := s.store.createAccount(Account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, Account)
}

func (s *apiServer) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	id, err := GetID(r)
	if err != nil {
		return err
	}
	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusNotFound, map[string]int{"Delete": id})
}

func (s *apiServer) TransferTo(w http.ResponseWriter, r *http.Request) error {
	var transferReq TransferRequest

	err := json.NewDecoder(r.Body).Decode(&transferReq)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return err
	}

	account, err := s.store.GetAccountbyID(transferReq.ToAccount)
	if err != nil {
		return err
	}
	account.Balance += transferReq.Amount

	error := s.store.UpdateAccount(account)
	if error != nil {
		return error
	}
	return WriteJSON(w, http.StatusOK, account.Balance)
}

func (s *apiServer) GetAccountbyID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		id, err := GetID(r)
		if err != nil {
			return err
		}
		account, err := s.store.GetAccountbyID(id)
		if err != nil {
			return err
		}

		return WriteJSON(w, http.StatusFound, account)
	}
	if r.Method == "DELETE" {
		return s.DeleteUser(w, r)
	}
	return fmt.Errorf("Method Not Allowed %s", r.Method)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJBY2NvdW50TnVtYmVyIjo2MDI5LCJleHBpcmVzQXQiOjE1MDAwfQ.zH7b_DqZbZgwdxRbqhnvp5660fbrM_LcMxumhgXixQAsXD9MH8H1-w2VTGN0LdPe1TBOmQBWqFN986otYal3dA
func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}

	secret := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func withJWTAuth(handlerFunc http.HandlerFunc, s storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")

		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			permissionDenied(w)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}
		userID, err := GetID(r)
		if err != nil {
			permissionDenied(w)
			return
		}
		account, err := s.GetAccountbyID(userID)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if account.Number != int(claims["accountNumber"].(float64)) {
			permissionDenied(w)
			return
		}

		handlerFunc(w, r)
	}
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, apiError{Error: "permission denied"})
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

type apiErrorHandler func(w http.ResponseWriter, r *http.Request) error

type apiError struct {
	Error string
}

func httpHandler(Req apiErrorHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := Req(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		}
	}

}
func GetID(r *http.Request) (int, error) {
	idstr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idstr)
	if err != nil {
		return id, fmt.Errorf("Invalid ID given %s", idstr)
	}
	return id, nil
}
