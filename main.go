package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/briannkhata/go-fiber-postgres/models"
	"github.com/briannkhata/go-fiber-postgres/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {

	book := Book{}
	err := context.BodyParser(&book)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "could not create book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book has been added"})
	return nil

}

func (r *Repository) GetBooks(context *fiber.Ctx) error {

	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"message": "could not get books"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "books fetched successfully",
		"data":    bookModels})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {

	bookModel := &[]models.Books{}
	id := context.Params("id")

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(bookModel, id).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not delete book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "books deleted successfully"})

	return nil

}

func (r *Repository) GetBookById(context *fiber.Ctx) error {

	id := context.Params("id")
	bookModel := &models.Books{}

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"message": "id cannot be empty"})
		return nil
	}
	fmt.Println("Id is ", id)

	err := r.DB.Where("id= ?",id).First(bookModel).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"message": "could not get a book"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "book fetched successfully",
		"data":    bookModel})

	return nil

}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookById)
	api.Get("/create_books", r.CreateBook)
	api.Get("/books", r.GetBooks)

}

func main() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate the db")
	}

	r := Repository{
		DB: db,
	}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8001")
}
