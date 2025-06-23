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
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateTask(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user_id").(string)

	var task models.Task

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		utils.ResponseWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	task.UserID = userId
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	collection := config.MongoClient.Database("task_db").Collection("tasks")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if t, err := collection.InsertOne(ctx, task); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			utils.ResponseWithError(w, http.StatusConflict, "task already exists")
		} else {
			utils.ResponseWithError(w, http.StatusInternalServerError, "failed to create task")
		}
		return
	} else {
		task.ID = t.InsertedID.(primitive.ObjectID)

		utils.ResponseWithJson(w, http.StatusCreated, task)
	}

}

func ListAllTask(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.Context().Value("user_id").(string)

	collection := config.MongoClient.Database("task_db").Collection("tasks")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{"user_id": userIdStr})
	if err != nil {
		utils.ResponseWithError(w, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}
	defer cursor.Close(ctx)

	var allTasks []models.Task
	if err = cursor.All(ctx, &allTasks); err != nil {
		utils.ResponseWithError(w, http.StatusInternalServerError, "Failed to decode tasks")
		return
	}

	utils.ResponseWithJson(w, http.StatusOK, allTasks)
}

func UpdateTask(w http.ResponseWriter, r *http.Request) {
	// userIdstr := r.Context().Value("user_id").(string)
	var UpdateReq models.Task

	if err := json.NewDecoder(r.Body).Decode(&UpdateReq); err != nil {
		utils.ResponseWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	updateField := bson.M{}

	if UpdateReq.Title != "" {
		updateField["title"] = UpdateReq.Title
	}
	if UpdateReq.Description != "" {
		updateField["description"] = UpdateReq.Description
	}
	if UpdateReq.Status != "" {
		updateField["status"] = UpdateReq.Status
	}
	if UpdateReq.Priority != "" {
		updateField["priority"] = UpdateReq.Priority
	}

	updateField["updated_at"] = time.Now()

	if len(updateField) < 2 {
		utils.ResponseWithError(w, http.StatusBadRequest, "no update request provided")
		return
	}

	collection := config.MongoClient.Database("task_db").Collection("tasks")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{"$set": updateField}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": UpdateReq.ID}, update)

	if err != nil {
		utils.ResponseWithError(w, http.StatusInternalServerError, "failed to update the task")
		return
	}
	if result.MatchedCount == 0 {
		utils.ResponseWithError(w, http.StatusNotFound, "Task not found")
		return
	}

	utils.ResponseWithJson(w, http.StatusOK, "task update sucessfully")

}

func DeletedTask(w http.ResponseWriter, r *http.Request) {
       
}
