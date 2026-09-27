package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gbplantwiki/gbplantwiki/internal/dto"
)

// The garden page date picker submits "YYYY-MM-DD"; binding must accept it
// instead of rejecting the save for not being RFC3339.
func TestReminderCreateRequestBindsPlainDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/t", func(c *gin.Context) {
		var req dto.ReminderCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"date": time.Time(req.RemindDate).Format("2006-01-02"), "garden_id": req.GardenID})
	})

	body := strings.NewReader(`{"task_title":"浇水","remind_date":"2026-09-27","garden_id":3}`)
	req := httptest.NewRequest(http.MethodPost, "/t", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("plain date should bind, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "2026-09-27") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

// A missing remind_date must still be rejected by `required` even though the
// field is a custom JSONDate type now.
func TestReminderCreateRequestRequiresDate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/t", func(c *gin.Context) {
		var req dto.ReminderCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/t", strings.NewReader(`{"task_title":"浇水"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing remind_date should fail required validation, got %d", w.Code)
	}
}
