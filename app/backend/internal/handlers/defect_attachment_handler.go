package handlers

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Quasar777/buildefect/app/backend/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type DefectAttachmentHandler struct {
	db *gorm.DB
}

func NewDefectAttachmentHandler(db *gorm.DB) *DefectAttachmentHandler {
	return &DefectAttachmentHandler{db: db}
}

func (h *DefectAttachmentHandler) UploadDefectAttachment(c *fiber.Ctx) error {
	defectID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid defect id"})
	}

	var defect models.Defect
	if err := h.db.First(&defect, defectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "defect not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// getting file (formdata)
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file required"})
	}

	// generating filename with random prefix + filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	path := fmt.Sprintf("internal/uploads/defect_attachments/%s", filename)

	// saving file locally
	if err := c.SaveFile(file, path); err != nil {
		fmt.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save file"})
	}

	// saving file in db
	attachment := models.DefectAttachment{
		DefectID: uint(defectID),
		URL:      path,
	}
	if err := h.db.Create(&attachment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to save attachment"})
	}

	return c.Status(fiber.StatusCreated).JSON(attachment)
}

func (h *DefectAttachmentHandler) GetDefectAttachments(c *fiber.Ctx) error {
    defectID, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid defect id"})
    }

    var defect models.Defect
    if err := h.db.First(&defect, defectID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "defect not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
    }

    attachments := []models.DefectAttachment{}
    if err := h.db.Where("defect_id = ?", defectID).Find(&attachments).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get attachments"})
    }

    return c.Status(fiber.StatusOK).JSON(attachments)
}

func (h *DefectAttachmentHandler) GetDefectAttachment(c *fiber.Ctx) error {
    attachmentID, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid attachment id"})
    }

    var attachment models.DefectAttachment
    if err := h.db.First(&attachment, attachmentID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "attachment not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
    }

    return c.Status(fiber.StatusOK).JSON(attachment)
}

func (h *DefectAttachmentHandler) DeleteDefectAttachment(c *fiber.Ctx) error {
	attachmentID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid attachment id"})
	}

	var attachment models.DefectAttachment
	result := h.db.First(&attachment, attachmentID)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "attachment not found"})
	}
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "database error"})
	}

	// deleting file from disk
	if err := os.Remove(attachment.URL); err != nil {
		fmt.Println("failed to remove file:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to remove file"})
	}

	// delete file from db
	if err := h.db.Delete(&attachment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete attachment"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "attachment deleted successfully"})
}
