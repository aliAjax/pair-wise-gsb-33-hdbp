package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/gbplantwiki/gbplantwiki/internal/model"
)

// CareReminderRepository handles persistence of care reminders.
type CareReminderRepository struct {
	db *gorm.DB
}

// NewCareReminderRepository creates a CareReminderRepository.
func NewCareReminderRepository(db *gorm.DB) *CareReminderRepository {
	return &CareReminderRepository{db: db}
}

// Create inserts a reminder.
func (r *CareReminderRepository) Create(m *model.CareReminder) error {
	return r.db.Create(m).Error
}

// FindByID locates a reminder by id.
func (r *CareReminderRepository) FindByID(id uint) (*model.CareReminder, error) {
	var m model.CareReminder
	if err := r.db.First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

// Update persists a reminder.
func (r *CareReminderRepository) Update(m *model.CareReminder) error {
	return r.db.Save(m).Error
}

// Delete removes a reminder.
func (r *CareReminderRepository) Delete(id uint) error {
	res := r.db.Delete(&model.CareReminder{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByUser returns reminders for a user with optional status and location
// filters. The location filter joins the linked garden item so reminders
// follow the plant: whatever location the plant currently sits in decides
// whether its reminders match.
func (r *CareReminderRepository) ListByUser(userID uint, status string, locationID uint) ([]model.CareReminder, error) {
	var items []model.CareReminder
	q := r.db.Model(&model.CareReminder{}).Where("care_reminders.user_id = ?", userID)
	if status != "" {
		q = q.Where("care_reminders.status = ?", status)
	}
	if locationID > 0 {
		q = q.Select("care_reminders.*").
			Joins("JOIN user_gardens ON user_gardens.id = care_reminders.garden_id AND user_gardens.location_id = ?", locationID)
	}
	if err := q.Order("care_reminders.remind_date ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ListByMonth returns reminders for a user within a month of a given year.
func (r *CareReminderRepository) ListByMonth(userID uint, year, month int) ([]model.CareReminder, error) {
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)
	var items []model.CareReminder
	if err := r.db.Where("user_id = ? AND remind_date >= ? AND remind_date < ?", userID, start, end).
		Order("remind_date ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// MarkOverdue flips pending reminders whose date has passed to overdue.
func (r *CareReminderRepository) MarkOverdue(userID uint) (int64, error) {
	res := r.db.Model(&model.CareReminder{}).
		Where("user_id = ? AND status = ? AND remind_date < ?", userID, model.ReminderPending, time.Now()).
		Update("status", model.ReminderOverdue)
	return res.RowsAffected, res.Error
}
