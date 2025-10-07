package handlers

import (
	"errors"
	"strconv"
	"time"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DefectHandler struct {
	db *gorm.DB
}

func NewDefectHandler(db *gorm.DB) *DefectHandler {
	return &DefectHandler{db: db}
}

// DTOs
type CreateDefectRequest struct {
	BuildingID          uint   `json:"building_id"`
	Title               string `json:"title"`
	Description         string `json:"description"`
	Priority            string `json:"priority"`               // low, medium, high
	ResponsiblePersonID *uint  `json:"responsible_person_id"`  // optional
	Deadline            string `json:"deadline"`               // RFC3339 string, optional
	Status              string `json:"status"`                 // optional, default "new"
}

type SimpleUser struct {
	ID       uint   `json:"id"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	Role     string `json:"role"`
}

type SimpleBuilding struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Stage   string `json:"stage"`
}

type DefectResponse struct {
	ID                  uint           `json:"id"`
	BuildingID          uint           `json:"building_id"`
	Building            SimpleBuilding `json:"building"`
	CreatedAt           time.Time      `json:"created_at"`
	CreatedByPersonID   uint           `json:"created_by_person_id"`
	CreatedBy           SimpleUser     `json:"created_by"`
	UpdatedAt           time.Time      `json:"updated_at"`
	UpdatedByPersonID   uint           `json:"updated_by_person_id"`
	Title               string         `json:"title"`
	Description         string         `json:"description"`
	Priority            string         `json:"priority"`
	ResponsiblePersonID *uint          `json:"responsible_person_id,omitempty"`
	Responsible         *SimpleUser    `json:"responsible,omitempty"`
	Deadline            *time.Time     `json:"deadline,omitempty"`
	Status              string         `json:"status"`
}

func toSimpleUser(u models.User) SimpleUser {
	return SimpleUser{ID: u.ID, Login: u.Login, Name: u.Name, LastName: u.LastName, Role: u.Role}
}

func toSimpleBuilding(b models.Building) SimpleBuilding {
	return SimpleBuilding{ID: b.ID, Name: b.Name, Address: b.Address, Stage: b.Stage}
}

func toDefectResponse(d models.Defect) DefectResponse {
	var resp DefectResponse
	resp.ID = d.ID
	resp.BuildingID = d.BuildingID
	resp.Building = toSimpleBuilding(d.Building)
	resp.CreatedAt = d.CreatedAt
	resp.CreatedByPersonID = d.CreatedByPersonID
	resp.CreatedBy = toSimpleUser(d.CreatedBy)
	resp.UpdatedAt = d.UpdatedAt
	resp.UpdatedByPersonID = d.UpdatedByPersonID
	resp.Title = d.Title
	resp.Description = d.Description
	resp.Priority = d.Priority
	resp.ResponsiblePersonID = &d.ResponsiblePersonID
	resp.Status = d.Status
	// Responsible may be empty zero value
	if d.Responsible.ID != 0 {
		tmp := toSimpleUser(d.Responsible)
		resp.Responsible = &tmp
	}
	// Deadline zero-time -> omit
	if !d.Deadline.IsZero() {
		tmp := d.Deadline
		resp.Deadline = &tmp
	}
	return resp
}

func (h *DefectHandler) CreateDefect(c *fiber.Ctx) error {
	var req CreateDefectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if req.BuildingID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "building_id is required"})
	}
	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title is required"})
	}

	// get current user from context (set by JWT middleware)
	uidRaw := c.Locals("user_id")
	if uidRaw == nil {
		// require auth for creation
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}
	createdByID, ok := uidRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in context"})
	}

	// parse deadline if provided
	var deadline time.Time
	if req.Deadline != "" {
		t, err := time.Parse(time.DateTime, req.Deadline) // DateTime Layout = "2006-01-02 15:04:05"
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid deadline format, use '2000-01-02 00:00:00'"})
		}
		deadline = t
	}

	// wrap in transaction: проверим связанные сущности и создадим дефект
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// check building exists
		var building models.Building
		if err := tx.First(&building, req.BuildingID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return fiber.NewError(fiber.StatusBadRequest, "building not found")
			}
			return err
		}

		// check responsible if provided
		if req.ResponsiblePersonID != nil {
			var resp models.User
			if err := tx.First(&resp, *req.ResponsiblePersonID).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return fiber.NewError(fiber.StatusBadRequest, "responsible person not found")
				}
				return err
			}
		}

		// default status
		status := req.Status
		if status == "" {
			status = "new"
		}

		defect := models.Defect{
			BuildingID:          req.BuildingID,
			CreatedByPersonID:   createdByID,
			UpdatedByPersonID:   createdByID,
			Title:               req.Title,
			Description:         req.Description,
			Priority:            req.Priority,
			ResponsiblePersonID: 0,
			Deadline:            deadline,
			Status:              status,
		}
		if req.ResponsiblePersonID != nil {
			defect.ResponsiblePersonID = *req.ResponsiblePersonID
		}

		if err := tx.Create(&defect).Error; err != nil {
			return err
		}

		// preload relations to return full response
		if err := tx.Preload("Building").Preload("CreatedBy").Preload("Responsible").First(&defect, defect.ID).Error; err != nil {
			return err
		}
		// attach defect to context for outer scope
		// but we will simply set it in closure variable by pointer if needed
		c.Locals("created_defect", defect) // optional: for middleware or logging
		return nil
	})

	if err != nil {
		// if Transaction returned fiber.Error, we can extract code/message
		if fe, ok := err.(*fiber.Error); ok {
			return c.Status(fe.Code).JSON(fiber.Map{"error": fe.Message})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// read created defect from context (set above)
	cdRaw := c.Locals("created_defect")
	if cdRaw == nil {
		// shouldn't happen but just in case
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load created defect"})
	}
	createdDefect, ok := cdRaw.(models.Defect)
	if !ok {
		// if you're using pointer set, handle accordingly
		if ptr, ok2 := cdRaw.(*models.Defect); ok2 {
			createdDefect = *ptr
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load created defect"})
		}
	}

	resp := toDefectResponse(createdDefect)
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *DefectHandler) GetDefects(c *fiber.Ctx) error {
	q := h.db.Preload("Building").Preload("CreatedBy").Preload("Responsible")

	// optional filters: status, building id, responsible person id
	if s := c.Query("status"); s != "" {
		q = q.Where("status = ?", s)
	}
	if b := c.Query("building_id"); b != "" {
		if bid, err := strconv.ParseUint(b, 10, 64); err == nil {
			q = q.Where("building_id = ?", uint(bid))
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid building_id"})
		}
	}
	if r := c.Query("responsible_id"); r != "" {
		if rid, err := strconv.ParseUint(r, 10, 64); err == nil {
			q = q.Where("responsible_person_id = ?", uint(rid))
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid responsible_id"})
		}
	}

	// pagination
	limit := 100
	if l := c.Query("limit"); l != "" {
		if li, err := strconv.Atoi(l); err == nil && li > 0 {
			limit = li
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid limit"})
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if oi, err := strconv.Atoi(o); err == nil && oi >= 0 {
			offset = oi
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid offset"})
		}
	}
	q = q.Limit(limit).Offset(offset)

	var defects []models.Defect
	if err := q.Find(&defects).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	resp := make([]DefectResponse, 0, len(defects))
	for _, d := range defects {
		resp = append(resp, toDefectResponse(d))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *DefectHandler) GetDefect(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var defect models.Defect
	result := h.db.Preload("Building").Preload("CreatedBy").Preload("Responsible").First(&defect, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "defect not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	return c.Status(fiber.StatusOK).JSON(toDefectResponse(defect))
}

// DeleteDefect — удаляет дефект по id
func (h *DefectHandler) DeleteDefect(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id",
		})
	}

	var defect models.Defect
	result := h.db.First(&defect, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "defect not found",
		})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	}

	// Пока что оставлю удаление чисто дефекта. Связанные с ним сущности удалять надо вручную.
	// TODO: обернуть в транзацкию для удаления связанных сущностей
	if err := h.db.Delete(&defect).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete defect",
		})
	}

	return c.Status(fiber.StatusOK).SendString("Successfully deleted defect with id " + strconv.Itoa(int(defect.ID)))
}