package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// ErrorResponse represents json response for errors
type ErrorResponse struct {
	Error string `json:"error"`
}

// Session used for storing data in mongodb
type Session struct {
	Id           uuid.UUID `bson:"_id"`
	RefreshToken []byte    `bson:"refreshToken,minsize"`
}

// TokenPair represents json response for getToken and refreshToken
type TokenPair struct {
	AccessToken  string `json:"accessToken"`  // jwt token
	RefreshToken string `json:"refreshToken"` // base64 encoded jwt token
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
}

func writeJSONError(w http.ResponseWriter, err error, code int) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	writeJSON(w, ErrorResponse{err.Error()})
}

// getTokenHandler takes uuid parameter from URL query and response with TokenPair if uuid is valid
func getTokenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id, err := uuid.Parse(r.URL.Query().Get("uuid"))
	if err != nil {
		writeJSONError(w, err, http.StatusBadRequest)
		return
	}

	at, rt, err := generateTokenPair(id)
	if err != nil {
		writeJSONError(w, err, http.StatusInternalServerError)
	}

	writeJSON(w, TokenPair{at, rt})
}

// refreshTokenHandler takes refreshToken from URL query and returns a new TokenPair.
// If refreshToken is invalid, returns ErrorResponse
func refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	decodedToken, err := base64.StdEncoding.DecodeString(r.URL.Query().Get("refreshToken"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(string(decodedToken), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecretKey, nil
	})

	if err != nil {
		writeJSONError(w, err, http.StatusBadRequest)
		return
	}

	rawId, err := token.Claims.GetSubject()
	if err != nil || !token.Valid {
		writeJSONError(w, err, http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(rawId)
	if err != nil {
		writeJSONError(w, err, http.StatusBadRequest)
		return
	}

	var session Session
	fmt.Println(id)
	err = collection.FindOne(dbCtx, bson.D{{"_id", id}}).Decode(&session)

	if errors.Is(err, mongo.ErrNoDocuments) {
		writeJSONError(w, err, http.StatusBadRequest)
		return
	} else if err != nil {
		// TODO
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword(session.RefreshToken, decodedToken)
	if err != nil {
		writeJSONError(w, errors.New("refresh token has been expired or revoked"), http.StatusBadRequest)
		return
	}

	at, rt, err := generateTokenPair(id)
	if err != nil {
		writeJSONError(w, err, http.StatusBadRequest)
		return
	}

	writeJSON(w, TokenPair{at, rt})
}
