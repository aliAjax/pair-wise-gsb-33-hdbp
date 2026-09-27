package repository

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCareReminderListByUserWithLocation(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewCareReminderRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT care_reminders.*, COALESCE(user_gardens.garden_no, '') AS garden_no, COALESCE(user_gardens.nickname, '') AS plant_nickname, COALESCE(user_gardens.location, '') AS location FROM `care_reminders` LEFT JOIN user_gardens ON user_gardens.id = care_reminders.garden_id WHERE care_reminders.user_id = ? AND user_gardens.location = ? ORDER BY care_reminders.remind_date ASC")).
		WithArgs(2, "balcony").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "plant_species_id", "garden_id", "task_title", "remind_date", "frequency", "status", "created_at", "garden_no", "plant_nickname", "location"}).
			AddRow(9, 2, 4, 1, "给月季补充缓释肥", time.Date(2026, 9, 30, 0, 0, 0, 0, time.Local), "monthly", "pending", time.Now(), "G-0001", "月季", "balcony"))
	items, err := repo.ListByUser(2, "", "balcony")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	got := items[0]
	if got.GardenNo != "G-0001" || got.PlantNickname != "月季" || got.Location != "balcony" || got.GardenID != 1 {
		t.Fatalf("join fields not populated: %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserGardenDeleteIsSoftDelete(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserGardenRepository(db)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `user_gardens` SET `deleted_at`=? WHERE `user_gardens`.`id` = ? AND `user_gardens`.`deleted_at` IS NULL")).
		WithArgs(sqlmock.AnyArg(), 7).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := repo.Delete(7); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestUserGardenCountAllIncludesRemoved(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewUserGardenRepository(db)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `user_gardens` WHERE user_id = ?")).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(5))
	total, err := repo.CountAllByUser(2)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected 5, got %d", total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
