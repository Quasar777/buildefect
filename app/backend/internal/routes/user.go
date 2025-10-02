package routes

import (
	"errors"
	"strconv"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserResponse struct {
	ID       uint   `json:"id"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	Role     string `json:"role"`
}

func RegisterUserRoutes(app *fiber.App, db *gorm.DB) {
	app.Post("/api/users", func(c *fiber.Ctx) error { return CreateUser(c, db) })
	app.Get("/api/users", func(c *fiber.Ctx) error { return GetUsers(c, db) })
	app.Get("/api/users/:id", func(c *fiber.Ctx) error { return GetUser(c, db) })
	app.Put("/api/users/:id", func(c *fiber.Ctx) error { return UpdateUser(c, db) })
	app.Delete("/api/users/:id", func(c *fiber.Ctx) error { return DeleteUser(c, db) })
}


func CreateResponseUser(userModel models.User) UserResponse {
	return UserResponse{
		ID: userModel.ID,
		Login: userModel.Login,
		Name: userModel.Name,
		LastName: userModel.LastName,
		Role: userModel.Role,
	}
}

func CreateUser(c *fiber.Ctx, db *gorm.DB) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var cnt int64
	if err := db.Model(&models.User{}).Where("login = ?", user.Login).Count(&cnt).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}
	if cnt > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user with this login already exists"})
	}

	if err := db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateResponseUser(user))
}


func GetUsers(c *fiber.Ctx, db *gorm.DB) error {
	users := []models.User{}

	if err := db.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}
	resp := make([]UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, CreateResponseUser(u))
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}


func GetUser(c *fiber.Ctx, db *gorm.DB) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var user models.User
	result := db.First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	return c.Status(fiber.StatusOK).JSON(CreateResponseUser(user))
}


func UpdateUser(c *fiber.Ctx, db *gorm.DB) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var user models.User
	result := db.First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	}
	
	type UpdateUserReq struct {
		Name string `json:"name"`
		LastName string `json:"lastname"`
	}

	var req UpdateUserReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	

	if err := db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save"})
	}

	return c.Status(fiber.StatusOK).JSON(CreateResponseUser(user))

}

func DeleteUser(c *fiber.Ctx, db *gorm.DB) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var user models.User
	result := db.First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	}

	if err := db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted user with id " + strconv.Itoa(int(user.ID)) )
}
