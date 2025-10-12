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

type DefectHandler struct {
	db *gorm.DB
}

func NewDefectHandler(db *gorm.DB) *DefectHandler {
	return &DefectHandler{db: db}
}

// CreateDefectRequest описывает тело запроса для создания дефекта.
// swagger:model CreateDefectRequest
type CreateDefectRequest struct {
    // example: 1
    BuildingID          uint   `json:"building_id"`
    // example: Трещина в стене
    Title               string `json:"title"`
    // example: Описание дефекта...
    Description         string `json:"description"`
    // example: high
    Priority            string `json:"priority"`               // low, medium, high
    // example: 2
    ResponsiblePersonID *uint  `json:"responsible_person_id"`  // optional
    // example: 2025-12-31 23:59:59
    // deadline in layout "2006-01-02 15:04:05"
    Deadline            string `json:"deadline"`               // optional, format: "2006-01-02 15:04:05"
    // example: new
    Status              string `json:"status"`                 // optional, default "new"
}

// SimpleUser краткая модель пользователя для вложенных сущностей.
// swagger:model SimpleUser
type SimpleUser struct {
    ID       uint   `json:"id"`
    Login    string `json:"login"`
    Name     string `json:"name"`
    LastName string `json:"lastname"`
    Role     string `json:"role"`
}

// SimpleBuilding краткая модель здания.
// swagger:model SimpleBuilding
type SimpleBuilding struct {
    ID      uint   `json:"id"`
    Name    string `json:"name"`
    Address string `json:"address"`
    Stage   string `json:"stage"`
}

// DefectResponse описывает ответ для дефекта.
// swagger:model DefectResponse
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

// UpdateStatusReq описывает тело запроса для изменения статуса дефекта.
// swagger:model UpdateStatusReq
type UpdateStatusReq struct {
	// example: in_progress
    // Возможные значения: "new", "in_progress", "review", "closed" (зависит от вашей логики)
	Status string `json:"status"`
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

// CreateDefect creates a new defect.
// @Summary     Create defect
// @Description Create a defect. Requires authentication.
// @Tags        defects
// @Accept      json
// @Produce     json
// @Param       payload  body      CreateDefectRequest  true  "Defect payload"
// @Success     201      {object}  DefectResponse
// @Failure     400      {object}  common.ErrorResponse  "invalid request body or missing fields"
// @Failure     401      {object}  common.ErrorResponse  "unauthenticated"
// @Failure     500      {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/defects [post]
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

// GetDefects returns list of defects with optional filters.
// @Summary     List defects
// @Description Retrieve defects with optional filters and pagination
// @Tags        defects
// @Accept      json
// @Produce     json
// @Param       status         query     string  false  "Filter by status"
// @Param       building_id    query     int     false  "Filter by building id"
// @Param       responsible_id query     int     false  "Filter by responsible user id"
// @Param       limit          query     int     false  "Limit number of results (default 100)"
// @Param       offset         query     int     false  "Offset for pagination (default 0)"
// @Success     200  {array}   DefectResponse
// @Failure     400  {object}  common.ErrorResponse  "invalid query param"
// @Failure     500  {object}  common.ErrorResponse
// @Router      /api/defects [get]
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

// GetDefect returns defect by id.
// @Summary     Get defect by id
// @Description Retrieve defect by numeric id
// @Tags        defects
// @Accept      json
// @Produce     json
// @Param       id   path      int  true  "Defect ID"
// @Success     200  {object}  DefectResponse
// @Failure     400  {object}  common.ErrorResponse  "invalid id"
// @Failure     404  {object}  common.ErrorResponse  "defect not found"
// @Failure     500  {object}  common.ErrorResponse
// @Router      /api/defects/{id} [get]
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

// UpdateStatus changes defect status according to role permissions.
// @Summary     Update defect status
// @Description Change status of a defect. Allowed status transitions depend on user role (engineer/manager). Allowed status values: new, in_progress, review, closed.
// @Tags        defects
// @Accept      json
// @Produce     json
// @Param       id      path      int             true  "Defect ID"
// @Param       payload body      UpdateStatusReq  true  "New status"
// @Success     200     {object}  DefectResponse
// @Failure     400     {object}  common.ErrorResponse  "invalid id or request body or missing status"
// @Failure     401     {object}  common.ErrorResponse  "unauthenticated"
// @Failure     403     {object}  common.ErrorResponse  "insufficient permissions or role cannot set this status"
// @Failure     404     {object}  common.ErrorResponse  "defect not found"
// @Failure     500     {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router 		/api/defects/{id} [patch]
func (h *DefectHandler) UpdateStatus(c *fiber.Ctx) error {
	// parse id
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	// parse body
	var req UpdateStatusReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.Status == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "status is required"})
	}

	// get role and user id from context (set by JWTMiddleware)
	roleRaw := c.Locals("role")
	if roleRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}
	role, ok := roleRaw.(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid role in context"})
	}

	uidRaw := c.Locals("user_id")
	if uidRaw == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthenticated"})
	}
	uid, ok := uidRaw.(uint)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid user id in context"})
	}

	// define allowed targets per role
	allowed := map[string]struct{}{}
	switch role {
	case "engineer":
		allowed["in_progress"] = struct{}{}
		allowed["review"] = struct{}{}
	case "manager":
		allowed["in_progress"] = struct{}{}
		allowed["review"] = struct{}{}
		allowed["closed"] = struct{}{}
	default:
		// other roles can't change status
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "insufficient permissions"})
	}

	// check requested status is allowed
	if _, ok := allowed[req.Status]; !ok {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "role cannot set this status"})
	}

	// load defect
	var defect models.Defect
	res := h.db.First(&defect, id)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "defect not found"})
	}
	if res.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// update fields
	defect.Status = req.Status
	defect.UpdatedByPersonID = uid

	if err := h.db.Save(&defect).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save status"})
	}

	// reload with relations for response
	if err := h.db.Preload("Building").Preload("CreatedBy").Preload("Responsible").First(&defect, defect.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load defect"})
	}

	return c.Status(fiber.StatusOK).JSON(toDefectResponse(defect))
}


// DeleteDefect deletes defect by id.
// @Summary     Delete defect
// @Description Delete defect by id
// @Tags        defects
// @Accept      json
// @Produce     plain
// @Param       id   path      int  true  "Defect ID"
// @Success     200  {string}  string  "Successfully deleted defect with id {id}"
// @Failure     400  {object}  common.ErrorResponse
// @Failure     404  {object}  common.ErrorResponse
// @Failure     500  {object}  common.ErrorResponse
// @Security    BearerAuth
// @Router      /api/defects/{id} [delete]
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