package handlers

import (
	"GoBackend/config"
	"GoBackend/models"
	"GoBackend/utils"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)


type AuthRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest


	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Hashed the Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.ResponseWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}


	collection := config.MongoClient.Database("task_db").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var existing models.User
	err = collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		utils.ResponseWithError(w, http.StatusConflict, "User already exists")
		return
	}
	if err != mongo.ErrNoDocuments {
		utils.ResponseWithError(w, http.StatusInternalServerError, "Database error")
		return
	}


	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}


	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		utils.ResponseWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	utils.ResponseWithJson(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest

	// Decode json Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseWithError(w, http.StatusBadRequest, "Invalid Request")
		return
	}



	collection := config.MongoClient.Database("task_db").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User

	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		utils.ResponseWithError(w, http.StatusUnauthorized, "User not found")
		return
	} else if err != nil {
		utils.ResponseWithError(w, http.StatusInternalServerError, "Database Error")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))

	if err != nil {
		utils.ResponseWithError(w, http.StatusUnauthorized, "Invalid Credentials")
		return
	}

	token, err := utils.GenerateJwt(user.ID.Hex())
	//  fmt.Println(user.ID.ObjectID[])

	if err != nil {
		utils.ResponseWithError(w, http.StatusInternalServerError, "Faild to genereate token")
		return
	}

	utils.ResponseWithJson(w, http.StatusOK, map[string]interface{}{
		"message": "Login Succesful",
		"token":   token,
		"user":    user,
	})

}
