package main

import (
	"fmt"
	"log"
	"os"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v2"
	swagger "github.com/gofiber/swagger"
	html "github.com/gofiber/template/html/v2"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	_ "github.com/prachaya-orr/fiber-api/docs"
)

func checkMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)

	if claims["role"] != "admin" {
		return fiber.ErrUnauthorized
	}

	return c.Next()
}

// @title Book API
// @description This is a sample server for a book API.
// @version 1.0
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	},
	)

	app.Get("/swagger/*", swagger.HandlerDefault) // default

	app.Post("/login", login)

	// JWT Middleware
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
	}))
	app.Use(checkMiddleware)

	app.Get("/books", getBooks)
	app.Get("/books/:id", getBookById)
	app.Post("/books", createBook)
	app.Patch("/books/:id", updateBook)
	app.Delete("/books/:id", deleteBook)

	app.Post("/upload", uploadFile)
	app.Get("/test-html", testHTML)

	app.Get("/config", getEnv)

	app.Listen(":8080")
}

func uploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	fmt.Println(file, err)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}
	c.SaveFile(file, "./uploads/"+file.Filename)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendString("File uploaded successfully!")
}

func testHTML(c *fiber.Ctx) error {
	return c.Render("index", fiber.Map{
		"Title": "Hello, World!",
		"Name":  "Go Learner",
	})
}

func getEnv(c *fiber.Ctx) error {

	return c.JSON(fiber.Map{
		"SECRET": os.Getenv("SECRET"),
	})
}

// Dummy user for example\

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

var memberUser = User{
	Email:    "user@example.com",
	Password: "password123",
}

func login(c *fiber.Ctx) error {
	user := new(User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	// user, pass ไม่ตรง =  Unauthorized
	if user.Email != memberUser.Email || user.Password != memberUser.Password {
		return fiber.ErrUnauthorized
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = user.Email
	claims["role"] = "admin"
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	//Generate encoded token and send it as response
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"message": "Login success",
		"token":   t,
	})
}
