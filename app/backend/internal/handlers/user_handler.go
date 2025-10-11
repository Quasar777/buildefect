package handlers

import (
	"errors"
	"strconv"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/Quasar777/buildefect/app/backend/internal/common"
)

var _ = common.ErrorResponse{} // костыль для swagger: без этой строчки будет ../common imported and not used

type UserHandler struct {
    db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
    return &UserHandler{db: db}
}

// CreateUserRequest represents request body to create a user.
// swagger:model CreateUserRequest
type CreateUserRequest struct {
    // example: ivan123
    Login    string `json:"login"`
    // example: passw0rd
    Password string `json:"password"`
    // example: Ivan
    Name     string `json:"name"`
    // example: Ivanov
    LastName string `json:"lastname"`
    // example: observer
    Role     string `json:"role"`
}

// UserResponse represents user data returned by API.
// swagger:model UserResponse
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

// UpdateUserRequest defines input for updating user.
// swagger:model UpdateUserRequest
type UpdateUserRequest struct {
    // example: Ivan
    Name     string `json:"name"`
    // example: Ivanov
    LastName string `json:"lastname"`
}

// CreateUser creates a new user.
// @Summary     Create a new user
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       user  body      CreateUserRequest  true  "New user data"
// @Success     201   {object}  UserResponse
// @Failure     400   {object}  common.ErrorResponse
// @Failure     409   {object}  common.ErrorResponse
// @Failure     500   {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
    }

    if req.Login == "" || req.Password == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "login and password are required"})
    }

    // проверка уникальности (уникальность также проверяется через тег unique в модели user)
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

// GetUsers returns list of users.
// @Summary     Get users
// @Description Retrieve list of all users
// @Tags        users
// @Accept      json
// @Produce     json
// @Success     200   {array}   UserResponse
// @Failure     500   {object}  common.ErrorResponse
// @Router      /api/users [get]
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

// GetUser returns a user by id.
// @Summary     Get user by id
// @Description Retrieve user by numeric id
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id   path      int  true  "User ID"
// @Success     200  {object}  UserResponse
// @Failure     400  {object}  common.ErrorResponse  "invalid id"
// @Failure     404  {object}  common.ErrorResponse  "user not found"
// @Failure     500  {object}  common.ErrorResponse
// @Router      /api/users/{id} [get]
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

// UpdateUser updates user's name/lastname.
// @Summary     Update user
// @Description Update user's name and/or lastname
// @Tags        users
// @Accept      json
// @Produce     json
// @Param       id    path      int                true  "User ID"
// @Param       data  body      UpdateUserRequest  true  "Update payload"
// @Success     200   {object}  UserResponse
// @Failure     400   {object}  common.ErrorResponse
// @Failure     404   {object}  common.ErrorResponse
// @Failure     500   {object}  common.ErrorResponse
// @Router      /api/users/{id} [patch]
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

// DeleteUser deletes user by id.
// @Summary     Delete user
// @Description Delete user by numeric id
// @Tags        users
// @Accept      json
// @Produce     plain
// @Param       id   path      int  true  "User ID"
// @Success     200  {string}  string  "Successfully deleted user with id {id}"
// @Failure     400  {object}  common.ErrorResponse
// @Failure     404  {object}  common.ErrorResponse
// @Failure     500  {object}  common.ErrorResponse
// @Router      /api/users/{id} [delete]
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