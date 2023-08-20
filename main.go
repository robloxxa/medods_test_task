package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	ACCESS_TOKEN_EXP  = 15 * time.Minute
	REFRESH_TOKEN_EXP = 60 * time.Minute
)

var (
	collection   *mongo.Collection
	dbCtx        = context.TODO()
	jwtSecretKey = []byte("SecretKey")
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("Please set MONGODB_URI environment variable")
	}

	jwtSecretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtSecretKey) < 0 {
		log.Fatal("Please set JWT_SECRET_KEY environment variable")
	}

	client, err := mongo.Connect(dbCtx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(dbCtx, nil)
	if err != nil {
		panic(err)
	}

	log.Println("Connected to the database")

	defer func() {
		if err := client.Disconnect(dbCtx); err != nil {
			log.Fatal(err)
		}
	}()

	collection = client.Database("users").Collection("sessions")

	http.HandleFunc("/getToken", getTokenHandler)
	http.HandleFunc("/refreshToken", refreshTokenHandler)

	log.Println("Running on port 9000")
	err = http.ListenAndServe(":9000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

// generateRefreshToken makes a new temporary jwt token with a UUID subject, and stores its bcrypted hash in a database.
// Returns base64 encoded jwt string
func generateRefreshToken(id uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = id.String()
	claims["exp"] = time.Now().UTC().Add(REFRESH_TOKEN_EXP).Unix()

	signedToken, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	bcryptedToken, err := bcrypt.GenerateFromPassword([]byte(signedToken), 14)
	if err != nil {
		return "", err
	}

	_, err = collection.ReplaceOne(dbCtx, bson.D{{"_id", id}}, Session{Id: id, RefreshToken: bcryptedToken}, options.Replace().SetUpsert(true))
	if err != nil {
		log.Println(err)
		return "", err
	}

	c, err := collection.CountDocuments(dbCtx, bson.D{})
	if err != nil {
		return "", err
	}
	fmt.Println(c)
	return base64.StdEncoding.EncodeToString([]byte(signedToken)), nil
}

// generateAccessToken makes a new temporary jwt token with a UUID subject.
func generateAccessToken(id uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = id.String()
	claims["exp"] = time.Now().UTC().Add(ACCESS_TOKEN_EXP).Unix()

	signedToken, err := token.SignedString(jwtSecretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func generateTokenPair(id uuid.UUID) (string, string, error) {
	at, err := generateAccessToken(id)
	if err != nil {
		return "", "", err
	}

	rt, err := generateRefreshToken(id)
	if err != nil {
		return "", "", err
	}

	return at, rt, nil
}
