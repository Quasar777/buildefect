package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"github.com/Quasar777/buildefect/app/backend/internal/common"
)

var _ = common.ErrorResponse{} // костыль для swagger: без этой строчки будет ../common imported and not used


type CommentHandler struct {
	db *gorm.DB
}

func NewCommentHandler(db *gorm.DB) *CommentHandler {
	return &CommentHandler{db: db}
}

// CreateCommentRequest описывает тело запроса для создания комментария
// swagger:model CreateCommentRequest
type CreateCommentRequest struct {
    // example: 1
    DefectID uint   `json:"defect_id"`
    // example: Found a broken window on 3rd floor
    Text     string `json:"text"`
}

// CommentResponse описывает комментарий
// swagger:model CommentResponse
type CommentResponse struct {
    // example: 1
    ID        uint      `json:"id"`
    // example: 1
    DefectID  uint      `json:"defect_id"`
    // example: 2025-10-11T14:00:00Z
    CreatedAt time.Time `json:"created_at"`
    // example: 2
    CreatedBy uint      `json:"created_by"`
    // example: Broken glass needs replacement
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

// CreateComment создаёт новый комментарий к дефекту
// @Summary     Create a comment
// @Description Create a new comment for a specific defect. Requires authentication.
// @Tags        comments
// @Accept      json
// @Produce     json
// @Param       comment  body      CreateCommentRequest  true  "Comment payload"
// @Success     201  {object}  CommentResponse
// @Failure     400  {object}  common.ErrorResponse
// @Failure     404  {object}  common.ErrorResponse
// @Failure     500  {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/comments [post]
func (h *CommentHandler) CreateComment(c *fiber.Ctx) error {
	var req CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.DefectID == 0 || req.Text == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "defect_id and text are required"})
	}

	var defect models.Defect
	if err := h.db.First(&defect, req.DefectID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "defect not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

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

// GetComments возвращает список комментариев для дефекта
// @Summary     List comments
// @Description Get all comments for a specific defect
// @Tags        comments
// @Accept      json
// @Produce     json
// @Param       defect_id  query     int  true  "Defect ID"
// @Success     200  {array}   CommentResponse
// @Failure     400  {object}  common.ErrorResponse
// @Failure     404  {object}  common.ErrorResponse
// @Failure     500  {object}  common.ErrorResponse
// @Router      /api/comments [get]
func (h *CommentHandler) GetComments(c *fiber.Ctx) error {
	var comments []models.Comment

	defectID := c.Query("defect_id")
	query := h.db
	if defectID != "" {
		query = query.Where("defect_id = ?", defectID)
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "defect_id parameter is required. Send it in URL params"})
	}

	if err := query.Order("created_at desc").Find(&comments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	resp := make([]CommentResponse, 0, len(comments))
	for _, comment := range comments {
		resp = append(resp, CreateResponseComment(comment))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// GetComment возвращает конкретный комментарий по ID
// @Summary     Get a comment
// @Description Get a comment by ID
// @Tags        comments
// @Accept      json
// @Produce     json
// @Param       id  path  int  true  "Comment ID"
// @Success     200  {object}  CommentResponse
// @Failure     400  {object}  common.ErrorResponse
// @Failure     404  {object}  common.ErrorResponse
// @Failure     500  {object}  common.ErrorResponse
// @Router      /api/comments/{id} [get]
func (h *CommentHandler) GetComment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var comment models.Comment
	result := h.db.Preload("CreatedBy").Preload("Attachments").First(&comment, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "comment not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	resp := CreateResponseComment(comment)
	return c.Status(fiber.StatusOK).JSON(resp)
}

// DeleteComment удаляет комментарий по ID (только для пользователя с ролью "observer")
// @Summary     Delete a comment
// @Description Delete a comment by ID. Only users with role "observer" are allowed.
// @Tags        comments
// @Accept      json
// @Produce     json
// @Param       id  path  int  true  "Comment ID"
// @Success     200  {string}  string  "Successfully deleted comment with id {id}"
// @Failure     400  {object}  common.ErrorResponse
// @Failure     403  {object}  common.ErrorResponse
// @Failure     404  {object}  common.ErrorResponse
// @Failure     500  {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/comments/{id} [delete]
func (h *CommentHandler) DeleteComment(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	// checking if user role is "observer"
	role := c.Locals("role")
	if roleStr, ok := role.(string); !ok || roleStr != "observer" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
	}

	var comment models.Comment
	result := h.db.First(&comment, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "comment not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	if err := h.db.Delete(&comment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete comment"})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted comment with id " + strconv.Itoa(int(comment.ID)))
}