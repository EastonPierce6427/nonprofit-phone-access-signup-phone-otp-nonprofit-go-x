package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/infrai-examples/nonprofit-phone-access/internal/infrai"
	"github.com/infrai-examples/nonprofit-phone-access/internal/nonprofit"
)

type server struct {
	infrai *infrai.Client
}

type signupRequest struct {
	Phone string         `json:"phone"`
	Role  nonprofit.Role `json:"role"`
	Code  string         `json:"code,omitempty"`
}

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}

	s := &server{infrai: infrai.New(key)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /signup/start", s.startSignup)
	mux.HandleFunc("POST /signup/verify", s.verifySignup)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("nonprofit phone access listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *server) startSignup(w http.ResponseWriter, r *http.Request) {
	var input signupRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if input.Phone == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone is required"})
		return
	}
	if _, err := nonprofit.AccessFor(input.Role); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	sum := sha256.Sum256([]byte(string(input.Role) + ":" + input.Phone))
	receipt, err := s.infrai.SendSignupCode(r.Context(), input.Phone, hex.EncodeToString(sum[:]))
	if err != nil {
		writeInfraiError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":     "code_sent",
		"message_id": receipt.MessageID,
	})
}

func (s *server) verifySignup(w http.ResponseWriter, r *http.Request) {
	var input signupRequest
	if err := decodeJSON(r, &input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if input.Phone == "" || input.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "phone and code are required"})
		return
	}
	access, err := nonprofit.AccessFor(input.Role)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	identity, err := s.infrai.VerifySignupCode(r.Context(), input.Phone, input.Code)
	if err != nil {
		writeInfraiError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, struct {
		Status string           `json:"status"`
		UserID string           `json:"user_id"`
		Role   nonprofit.Role   `json:"role"`
		Access nonprofit.Access `json:"access"`
	}{Status: "verified", UserID: identity.UserID, Role: input.Role, Access: access})
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeInfraiError(w http.ResponseWriter, err error) {
	var apiErr *infrai.APIError
	if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
		writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Message, "code": apiErr.Code})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": "identity provider request failed"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode response: %v", err)
	}
}
