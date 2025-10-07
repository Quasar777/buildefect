package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)


type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

type CreateCommentRequest struct {
	DefectID uint   `json:"defect_id"`
	Text     string `json:"text"`
}

type CommentResponse struct {
	ID        uint      `json:"id"`
	DefectID  uint      `json:"defect_id"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy uint      `json:"created_by"`
	Text      string    `json:"text"`
}

func CreateResponseComment(comment models.Comment) CommentResponse {
	return CommentResponse{
		ID:        comment.ID,
		DefectID:  comment.DefectID,
		CreatedAt: comment.CreatedAt,
		CreatedBy: comment.CreatedByPersonID,
		Text:      comment.Text,
	}
}

func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	var req CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.DefectID == 0 || req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "defect_id and text are required"})
	}

	// проверяем, существует ли дефект
	var defect models.Defect
	if err := h.db.First(&defect, req.DefectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "defect not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// получаем user_id из JWT
	userID := c.Locals("user_id").(uint)

	comment := models.Comment{
		DefectID:          req.DefectID,
		Text:              req.Text,
		CreatedAt:         time.Now(),
		CreatedByPersonID: userID,
	}

	if err := h.db.Create(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create comment"})
	}

	return c.Status(fiber.StatusCreated).JSON(CreateResponseComment(comment))
}

// передаем сюда query-параметр defect_id, чтобы загрузить комментарии для определенного дефекта.
// Все существующие комментарии получить НЕЛЬЗЯ (а зачем вообще?)
func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	var comments []models.Comment

	// проверяем, передан ли query параметр defect_id
	defectID := c.Query("defect_id")
	query := h.db
	if defectID != "" {
		query = query.Where("defect_id = ?", defectID)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "defect_id parameter is required. Send it in URL params"})
	}

	// загружаем комментарии начиная с последнего
	if err := query.Order("created_at desc").Find(&comments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// формируем response
	resp := make([]CommentResponse, 0, len(comments))
	for _, comment := range comments {
		resp = append(resp, CreateResponseComment(comment))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *CommentHandler) GetComment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var comment models.Comment
	// подгружаем автора и вложения
	result := h.db.Preload("CreatedBy").Preload("Attachments").First(&comment, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "comment not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// формируем response
	resp := CreateResponseComment(comment)
	return c.Status(fiber.StatusOK).JSON(resp)
}

// Комментарии может удалять только пользователь с ролью observer
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	// Проверяем роль пользователя, которая хранится в c.Locals("role") после JWTMiddleware
	role := c.Locals("role")
	if roleStr, ok := role.(string); !ok || roleStr != "observer" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
	}

	// Ищем комментарий в БД
	var comment models.Comment
	result := h.db.First(&comment, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "comment not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// Удаляем комментарий
	if err := h.db.Delete(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete comment"})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted comment with id " + strconv.Itoa(int(comment.ID)))
}