package router

import (
	"log/slog"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/gbplantwiki/gbplantwiki/internal/config"
)

// Setup must not panic on route conflicts (static segments like /gardens/moves
// coexist with /gardens/:id) and must register every garden endpoint.
func TestSetupRegistersGardenRoutes(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	r := Setup(config.Load(), db, logger)

	registered := map[string]bool{}
	for _, ri := range r.Routes() {
		registered[ri.Method+" "+ri.Path] = true
	}
	want := []string{
		"GET /api/v1/gardens",
		"POST /api/v1/gardens",
		"GET /api/v1/gardens/locations",
		"PUT /api/v1/gardens/locations/:location",
		"GET /api/v1/gardens/moves",
		"POST /api/v1/gardens/moves",
		"PUT /api/v1/gardens/:id/location",
		"PUT /api/v1/gardens/:id/reminder",
		"DELETE /api/v1/gardens/:id",
	}
	for _, w := range want {
		if !registered[w] {
			t.Errorf("missing route %s", w)
		}
	}
}
