package handlers

import (
	dbmodels "bookswapper/models/database"
	"bookswapper/models/requests"
	"bookswapper/utils/password"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"time"
)

func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := &requests.User{}
		if err := c.BodyParser(data); err != nil {
			errorString := fmt.Sprintf("invalid json: %s", err.Error())
			return c.Status(400).JSON(fiber.Map{
				"status": errorString,
			})
		}
		var user dbmodels.User
		result := db.First(&user, "login = ?", data.Login)
		if result.Error != nil {
			errorString := fmt.Sprintf("failed to find user: %s", result.Error.Error())
			return c.Status(404).JSON(fiber.Map{
				"status": errorString,
			})
		}
		if !password.CheckPasswordHash(data.Password, user.PasswordHash) {
			return c.Status(401).JSON(fiber.Map{
				"status": "wrong password",
			})
		}

		claims := jwt.MapClaims{
			"login": user.Login,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signedToken, tokenErr := token.SignedString([]byte("bookswapper"))
		if tokenErr != nil {
			errorString := fmt.Sprintf("failed to sign token: %s", tokenErr.Error())
			return c.Status(500).JSON(fiber.Map{
				"status": errorString,
			})
		}
		return c.JSON(&fiber.Map{
			"token": signedToken,
		})
	}
}

func Register(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := &requests.User{}
		if reqErr := c.BodyParser(data); reqErr != nil {
			errorString := fmt.Sprintf("invalid json: %s", reqErr.Error())
			return c.Status(400).JSON(fiber.Map{
				"status": errorString,
			})
		}

		hashedPassword, hashErr := password.HashPassword(data.Password)
		if hashErr != nil {
			errorString := fmt.Sprintf("failed to hash password: %s", hashErr.Error())
			return c.Status(400).JSON(fiber.Map{
				"status": errorString,
			})
		}
		user := &dbmodels.User{
			Login:        data.Login,
			PasswordHash: hashedPassword,
			CreatedAt:    time.Now(),
		}
		result := db.Create(&user)
		if result.Error != nil {
			errorString := fmt.Sprintf("failed to create user: %s", result.Error.Error())
			return c.Status(400).JSON(fiber.Map{
				"status": errorString,
			})
		}
		return c.JSON(fiber.Map{
			"status": "user created",
		})
	}
}