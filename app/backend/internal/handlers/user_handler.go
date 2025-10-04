package handlers

import (
	"errors"
	"strconv"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
    db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
    return &UserHandler{db: db}
}

type CreateUserRequest struct {
    Login    string `json:"login"`
    Password string `json:"password"`
    Name     string `json:"name"`
    LastName string `json:"lastname"`
    Role     string `json:"role"`
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	Role     string `json:"role"`
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

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
    }

    if req.Login == "" || req.Password == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "login and password are required"})
    }

    // проверка уникальности
    var cnt int64
    if err := h.db.Model(&models.User{}).Where("login = ?", req.Login).Count(&cnt).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
    }
    if cnt > 0 {
        return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user already exists"})
    }

    // хешируем пароль
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
    }

    user := models.User{
        Login:        req.Login,
        Password: string(hash),
        Name:         req.Name,
        LastName:     req.LastName,
        Role:         req.Role,
    }

    if err := h.db.Create(&user).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
    }

    return c.Status(fiber.StatusCreated).JSON(CreateResponseUser(user))
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	users := []models.User{}

	if err := h.db.Find(&users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}
	resp := make([]UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, CreateResponseUser(u))
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var user models.User
	result := h.db.First(&user, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	return c.Status(fiber.StatusOK).JSON(CreateResponseUser(user))
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var user models.User
	result := h.db.First(&user, id)
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

	if err := h.db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save"})
	}

	return c.Status(fiber.StatusOK).JSON(CreateResponseUser(user))
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var user models.User
	result := h.db.First(&user, id)
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

	if err := h.db.Delete(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted user with id " + strconv.Itoa(int(user.ID)) )
}

//function for test
func (h *UserHandler) GetUserByCtx(c *fiber.Ctx) error {
    uidRaw := c.Locals("user_id")
    if uidRaw == nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
    }
    uid := uidRaw.(uint)
    var user models.User
    if err := h.db.First(&user, uid).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db error"})
    }
    return c.Status(fiber.StatusOK).JSON(CreateResponseUser(user))
}
