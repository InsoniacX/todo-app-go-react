package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Completed bool               `json:"completed" bson:"completed"`
	Body      string             `json:"body" bson:"body"`
}

var collection *mongo.Collection

func main() {

	if os.Getenv("ENV") != "production" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("No .env file found")
		}
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")
	if MONGODB_URI == "" {
		log.Fatal("MONGODB_URI is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("CONNECTED TO MONGODB")

	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	allowedOrigin := os.Getenv("CORS_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	if os.Getenv("ENV") != "production" {
		app.Use(cors.New(cors.Config{
			AllowOrigins: "http://localhost:5173",
			AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
			AllowHeaders: "Origin, Content-Type, Accept",
		}))
	}

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func getTodos(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var todos []Todo

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch todos"})
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to decode todo"})
		}
		todos = append(todos, todo)
	}

	if err := cursor.Err(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Cursor error"})
	}

	return c.JSON(todos)
}

func createTodo(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Todo body cannot be empty"})
	}

	todo.Completed = false

	insertResult, err := collection.InsertOne(ctx, todo)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create todo"})
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Params("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{"$set": bson.M{"completed": true}}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update todo"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.JSON(fiber.Map{"success": true})
}

func deleteTodo(c *fiber.Ctx) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	id := c.Params("id")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid todo ID"})
	}

	filter := bson.M{"_id": objectId}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete todo"})
	}

	if result.DeletedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Todo not found"})
	}

	return c.JSON(fiber.Map{"success": true})
}
