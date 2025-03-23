package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	//"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	//"gorm.io/driver/postgres"
	//"gorm.io/gorm"
	"errors"

	"github.com/google/uuid"
)

type UserService interface {
	Get(login string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByPhoneNumber(phone_number string) (*User, error)
	UpdateAuthToken(id uuid.UUID, token string) error
	ValidatePassword(username, password string) (*User, error)
	Save(*User) (*User, error)
	Create(user *User) error
}

type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Username     string    `json:"login" gorm:"unique;not null"`
	Email        string    `json:"email" gorm:"unique;not null"`
	PhoneNumber  string    `json:"phone_number"`
	PasswordHash string    `json:"-" gorm:"not null"`
	AuthToken    string    `json:"auth_token"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

var (
	NotFound = errors.New("key not found")
)

type UserHandler struct {
	Service UserService
}

func NewUserHandler(s UserService) *UserHandler {
	return &UserHandler{
		Service: s,
	}
}

func (h *UserHandler) RegisterRoutes(r fiber.Router) {
	userGroup := r.Group("/api/users")
	userGroup.Post("/", h.CreateUser)
	userGroup.Get("/:login", h.GetUserByLogin)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	//createdUser, err := h.UserService.Create(&user)
	//if err != nil {
	//	return c.Status(http.StatusBadRequest).JSON(fiber.Map{
	//		"error": err.Error(),
	//	})
	//}
	return c.SendStatus(fiber.StatusOK)
}

func (h *UserHandler) GetUserByLogin(c *fiber.Ctx) error {
	login := c.Params("login")
	user, err := h.Service.Get(login)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(err)
	}
	return c.JSON(user)
}

func main() {
	addr := flag.String("addr", ":8080", "HTTP server address")
	flag.Parse()

	//psqlInfo := "host=localhost user=postgres password=yourpassword dbname=legal_db port=5432 sslmode=disable"
	//db, err := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	//if err != nil {
	//    log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	//}

	//redisClient := redis.NewClient(&redis.Options{
	//    Addr:     "localhost:6379",
	//   Password: "",
	//    DB:       0,
	//})

	var dummyService UserService
	userHandler := NewUserHandler(dummyService)

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	app.Use(logger.New())
	app.Use(recover.New())

	userGroup := app.Group("/api/users")
	userHandler.RegisterRoutes(userGroup)

	log.Printf("Starting server on %s", *addr)
	if err := app.Listen(*addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
