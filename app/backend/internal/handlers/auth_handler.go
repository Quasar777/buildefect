package handlers

import (
	"strconv"
	"time"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"github.com/Quasar777/buildefect/app/backend/internal/common"
)

var _ = common.ErrorResponse{} // костыль для swagger: без этой строчки будет ../common imported and not used

type AuthHandler struct {
	db        *gorm.DB
	jwtSecret string
	ttl       time.Duration
}

func NewAuthHandler(db *gorm.DB, jwtSecret string, ttl time.Duration) *AuthHandler {
	return &AuthHandler{
		db: db, 
		jwtSecret: jwtSecret, 
		ttl: ttl,
	}
}


// RegisterRequest represents request to register a new user.
// swagger:model RegisterRequest
type RegisterRequest struct {
    // example: ivan123
    Login    string `json:"login"`
    // example: passw0rd
    Password string `json:"password"`
    // example: Ivan
    Name     string `json:"name"`
    // example: Ivanov
    LastName string `json:"lastname"`
}

// LoginRequest represents login request.
// swagger:model LoginRequest
type LoginRequest struct {
    // example: ivan123
    Login    string `json:"login"`
    // example: passw0rd
    Password string `json:"password"`
}

// TokenResponse represents JWT token response.
// swagger:model TokenResponse
type TokenResponse struct {
    // access token signed with server secret
    AccessToken string `json:"access_token"`
    // token type, usually "Bearer"
    TokenType   string `json:"token_type"`
    // seconds until token expiration
    ExpiresIn   int64  `json:"expires_in"`
}

// Register registers a new user (no token returned).
// @Summary     Register a user
// @Description Create a new user account (no JWT returned). Default role = "engineer".
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       payload  body      RegisterRequest  true  "Registration payload"
// @Success     201      {object}  UserResponse
// @Failure     400      {object}  common.ErrorResponse  "invalid request body or missing fields"
// @Failure     409      {object}  common.ErrorResponse  "user already exists"
// @Failure     500      {object}  common.ErrorResponse
// @Router      /api/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Login == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "login and password required"})
	}
	var cnt int64
	if err := h.db.Model(&models.User{}).Where("login = ?", req.Login).Count(&cnt).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}
	if cnt > 0 {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "user already exists"})
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	user := models.User{
		Login:        req.Login,
		Password: string(hash),
		Name:         req.Name,
		LastName:     req.LastName,
		Role:         "engineer",
	}

	if err := h.db.Create(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}

	// не возвращаем хэш, только публичную инфу
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":       user.ID,
		"login":    user.Login,
		"name":     user.Name,
		"lastname": user.LastName,
		"role":     user.Role,
	})
}

// Login authenticates user and returns JWT token.
// @Summary     Login and obtain JWT
// @Description Validate credentials and return access token with expiry seconds
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       payload  body      LoginRequest   true  "Login payload"
// @Success     200      {object}  TokenResponse
// @Failure     400      {object}  common.ErrorResponse  "invalid request body"
// @Failure     401      {object}  common.ErrorResponse  "invalid credentials"
// @Failure     500      {object}  common.ErrorResponse
// @Router      /api/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	var user models.User
	if err := h.db.Where("login = ?", req.Login).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	// creating jwt payload
	now := time.Now()
	exp := now.Add(h.ttl)
	claims := jwt.MapClaims{
		"sub":   strconv.FormatUint(uint64(user.ID), 10),
		"login": user.Login,
		"role":  user.Role,
		"iat":   now.Unix(),
		"exp":   exp.Unix(),
	}

	// generating and signing token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to sign token"})
	}

	return c.Status(fiber.StatusOK).JSON(TokenResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.ttl.Seconds()),
	})
}