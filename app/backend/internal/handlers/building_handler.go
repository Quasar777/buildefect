package handlers

import (
	"errors"
	"strconv"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type BuildingHandler struct {
	db *gorm.DB
}

func NewBuildingHandler(db *gorm.DB) *BuildingHandler {
	return &BuildingHandler{db: db}
}

type CreateBuildingRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Stage   string `json:"stage"`
}

type BuildingResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Stage   string `json:"stage"`
}

func toBuildingResponse(b models.Building) BuildingResponse {
	return BuildingResponse{
		ID:      b.ID,
		Name:    b.Name,
		Address: b.Address,
		Stage:   b.Stage,
	}
}

func (h *BuildingHandler) CreateBuilding(c *fiber.Ctx) error {
	var req CreateBuildingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	building := models.Building{
		Name:    req.Name,
		Address: req.Address,
		Stage:   req.Stage,
	}

	if err := h.db.Create(&building).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create building"})
	}

	return c.Status(fiber.StatusCreated).JSON(toBuildingResponse(building))
}

func (h *BuildingHandler) GetBuildings(c *fiber.Ctx) error {
	var buildings []models.Building

	if err := h.db.Find(&buildings).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	}

	resp := make([]BuildingResponse, 0, len(buildings))
	for _, b := range buildings {
		resp = append(resp, toBuildingResponse(b))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *BuildingHandler) GetBuilding(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var building models.Building
	result := h.db.First(&building, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "building not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	return c.Status(fiber.StatusOK).JSON(toBuildingResponse(building))
}

func (h *BuildingHandler) UpdateBuilding(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var building models.Building
	result := h.db.First(&building, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "building not found",
		})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	}

	// DTO для частичного обновления
	type UpdateBuildingReq struct {
		Name    string `json:"name"`
		Address string `json:"address"`
		Stage   string `json:"stage"`
	}

	var req UpdateBuildingReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.Name != "" {
		building.Name = req.Name
	}
	if req.Address != "" {
		building.Address = req.Address
	}
	if req.Stage != "" {
		building.Stage = req.Stage
	}

	if err := h.db.Save(&building).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save building"})
	}

	return c.Status(fiber.StatusOK).JSON(toBuildingResponse(building))
}

func (h *BuildingHandler) DeleteBuilding(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var building models.Building
	result := h.db.First(&building, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "building not found",
		})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	}

	if err := h.db.Delete(&building).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete building",
		})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted building with id " + strconv.Itoa(int(building.ID)))
}
