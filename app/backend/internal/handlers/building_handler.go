package handlers

import (
	"errors"
	"strconv"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"github.com/Quasar777/buildefect/app/backend/internal/common"
)

var _ = common.ErrorResponse{} // костыль для swagger: без этой строчки будет ../common imported and not used

type BuildingHandler struct {
	db *gorm.DB
}

func NewBuildingHandler(db *gorm.DB) *BuildingHandler {
	return &BuildingHandler{db: db}
}

// CreateBuildingRequest описывает тело запроса для создания здания.
// swagger:model CreateBuildingRequest
type CreateBuildingRequest struct {
    // example: Дом на Невском
    Name    string `json:"name"`
    // example: Невский пр., 1
    Address string `json:"address"`
    // example: построено
    Stage   string `json:"stage"`
}

// BuildingResponse описывает структуру ответа для здания.
// swagger:model BuildingResponse
type BuildingResponse struct {
    ID      uint   `json:"id"`
    Name    string `json:"name"`
    Address string `json:"address"`
    Stage   string `json:"stage"`
}

// UpdateBuildingRequest используется для частичного обновления здания.
// swagger:model UpdateBuildingRequest
type UpdateBuildingRequest struct {
    // example: Новый дом
    Name    string `json:"name"`
    // example: Новый адрес
    Address string `json:"address"`
    // example: в_строительстве
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

// CreateBuilding creates a new building.
// @Summary     Create building
// @Description Create a new building record
// @Tags        buildings
// @Accept      json
// @Produce     json
// @Param       payload body     CreateBuildingRequest true "Building payload"
// @Success     201     {object} BuildingResponse
// @Failure     400     {object} common.ErrorResponse "invalid request body or missing fields"
// @Failure     500     {object} common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/buildings [post]
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

// GetBuildings returns list of buildings.
// @Summary     List buildings
// @Description Retrieve list of all buildings
// @Tags        buildings
// @Accept      json
// @Produce     json
// @Success     200  {array}  BuildingResponse
// @Failure     500  {object} common.ErrorResponse
// @Router      /api/buildings [get]
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

// GetBuilding returns building by id.
// @Summary     Get building by id
// @Description Retrieve building by numeric id
// @Tags        buildings
// @Accept      json
// @Produce     json
// @Param       id   path      int  true  "Building ID"
// @Success     200  {object}  BuildingResponse
// @Failure     400  {object}  common.ErrorResponse "invalid id"
// @Failure     404  {object}  common.ErrorResponse "building not found"
// @Failure     500  {object}  common.ErrorResponse
// @Router      /api/buildings/{id} [get]
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

// UpdateBuilding updates building fields partially.
// @Summary     Update building
// @Description Partially update building (name/address/stage)
// @Tags        buildings
// @Accept      json
// @Produce     json
// @Param       id      path      int                   true  "Building ID"
// @Param       payload body      UpdateBuildingRequest  true  "Update payload"
// @Success     200     {object}  BuildingResponse
// @Failure     400     {object}  common.ErrorResponse
// @Failure     404     {object}  common.ErrorResponse
// @Failure     500     {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/buildings/{id} [patch]
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

	// TODO: Наверно можно убрать, если я вынес это вверху
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

// DeleteBuilding removes building by id.
// @Summary     Delete building
// @Description Delete building by id
// @Tags        buildings
// @Accept      json
// @Produce     plain
// @Param       id   path      int  true  "Building ID"
// @Success     200  {string}  string "Successfully deleted building with id {id}"
// @Failure     400  {object}  common.ErrorResponse
// @Failure     404  {object}  common.ErrorResponse
// @Failure     500  {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/buildings/{id} [delete]
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
