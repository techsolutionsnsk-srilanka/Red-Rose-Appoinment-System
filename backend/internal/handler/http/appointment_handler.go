package http

import (
	"net/http"
	"strconv"
	"time"

	"redrose/backend/internal/domain"
	"redrose/backend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	uc *usecase.AppointmentUseCase
}

func NewAppointmentHandler(uc *usecase.AppointmentUseCase) *AppointmentHandler {
	return &AppointmentHandler{uc: uc}
}

// CreateAppointment handles POST /api/appointments
func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	var req struct {
		CustomerName  string    `json:"customer_name" binding:"required"`
		CustomerEmail string    `json:"customer_email" binding:"required,email"`
		CustomerPhone string    `json:"customer_phone"`
		Service       string    `json:"service" binding:"required"`
		StartsAt      time.Time `json:"starts_at" binding:"required"`
		DurationMin   int       `json:"duration_min"`
		Notes         string    `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body: " + err.Error()})
		return
	}

	appt := &domain.Appointment{
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		Service:       req.Service,
		StartsAt:      req.StartsAt,
		DurationMin:   req.DurationMin,
		Notes:         req.Notes,
		Status:        domain.AppointmentStatusPending,
	}
	// If an authenticated admin created it, record their Clerk user ID.
	if uid, ok := c.Get("user_id"); ok {
		if s, ok := uid.(string); ok {
			appt.CreatedBy = s
		}
	}

	if err := h.uc.CreateAppointment(c.Request.Context(), appt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":       true,
		"message":       "Appointment booked successfully",
		"appointmentId": appt.ID,
		"data":          appt,
	})
}

// GetAppointment handles GET /api/appointments/:id
func (h *AppointmentHandler) GetAppointment(c *gin.Context) {
	appt, err := h.uc.GetAppointment(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if appt == nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Appointment not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": appt})
}

// ListAppointments handles GET /api/appointments
func (h *AppointmentHandler) ListAppointments(c *gin.Context) {
	filters := domain.AppointmentFilters{
		Status:        domain.AppointmentStatus(c.Query("status")),
		CustomerEmail: c.Query("customer_email"),
		Service:       c.Query("service"),
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.FromDate = t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filters.ToDate = t
		}
	}
	if v := c.Query("limit"); v != "" {
		filters.Limit, _ = strconv.Atoi(v)
	}
	if v := c.Query("offset"); v != "" {
		filters.Offset, _ = strconv.Atoi(v)
	}

	appts, err := h.uc.ListAppointments(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": appts, "count": len(appts)})
}

// UpdateStatus handles PUT /api/appointments/:id/status
func (h *AppointmentHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}
	if err := h.uc.UpdateStatus(c.Request.Context(), c.Param("id"), domain.AppointmentStatus(req.Status)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Status updated successfully"})
}

// UpdateAdminNotes handles PUT /api/appointments/:id/notes
func (h *AppointmentHandler) UpdateAdminNotes(c *gin.Context) {
	var req struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid request body"})
		return
	}
	if err := h.uc.UpdateAdminNotes(c.Request.Context(), c.Param("id"), req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Admin notes updated successfully"})
}

// DeleteAppointment handles DELETE /api/appointments/:id
func (h *AppointmentHandler) DeleteAppointment(c *gin.Context) {
	if err := h.uc.DeleteAppointment(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Appointment deleted successfully"})
}

// GetStats handles GET /api/appointments/stats
func (h *AppointmentHandler) GetStats(c *gin.Context) {
	stats, err := h.uc.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}
