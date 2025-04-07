package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Task struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Title   string             `bson:"title" json:"title"`
	Done    bool               `bson:"done" json:"done"`
	Created time.Time          `bson:"created" json:"created"`
}

var collection *mongo.Collection

func main() {
	// Conexión Mongo
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb+srv://admin:admin123@testdb.uyrq69i.mongodb.net/?retryWrites=true&w=majority&appName=TestDB"))
	if err != nil {
		panic(err)
	}
	collection = client.Database("todolist").Collection("tasks")

	router := gin.Default()

	// Sirve el HTML en la raíz
	router.Static("/static", "./static")
	router.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Endpoints REST
	router.GET("/tasks", getTasks)
	router.POST("/tasks", createTask)
	router.PUT("/tasks/:id", updateTask)
	router.DELETE("/tasks/:id", deleteTask)

	router.Run(":8080")
}

func getTasks(c *gin.Context) {
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cursor.Close(context.TODO())

	var tasks []Task
	if err = cursor.All(context.TODO(), &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func createTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task.Created = time.Now()
	task.Done = false

	result, err := collection.InsertOne(context.TODO(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	task.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusOK, task)
}

func updateTask(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var update bson.M
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func deleteTask(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	_, err = collection.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
