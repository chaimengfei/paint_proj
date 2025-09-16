package repository

import (
	"fmt"
	"testing"
	"time"

	"cmf/paint_proj/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// newTestDB returns an isolated in-memory sqlite DB configured with singular table names.
// It also creates the minimal tables needed by the repository queries.
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // ensures table name "address" / "user" to match explicit SQL in repository
		},
		SkipDefaultTransaction: true,
	})
	if err \!= nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}

	// Create minimal schema required by queries.
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY,
			nickname TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS address (
			id INTEGER PRIMARY KEY,
			user_id INTEGER NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			is_delete INTEGER NOT NULL DEFAULT 0
		);`,
	}
	for _, s := range sqls {
		if err := db.Exec(s).Error; err \!= nil {
			t.Fatalf("failed to create schema: %v", err)
		}
	}

	return db
}

func mustExec(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	if err := db.Exec(sql, args...).Error; err \!= nil {
		t.Fatalf("exec failed: %v; sql=%s; args=%v", err, sql, args)
	}
}

func insertUser(t *testing.T, db *gorm.DB, id int64, nickname string) {
	mustExec(t, db, `INSERT INTO user(id, nickname) VALUES(?, ?)`, id, nickname)
}

func insertAddress(t *testing.T, db *gorm.DB, id, userID int64, isDefault, isDelete int) {
	mustExec(t, db, `INSERT INTO address(id, user_id, is_default, is_delete) VALUES(?, ?, ?, ?)`,
		id, userID, isDefault, isDelete)
}

func TestGetAddressListByUser_NoFilters_ReturnsAllFromAddressOnly(t *testing.T) {
	db := newTestDB(t)
	// Seed address rows; user_name should be empty string when no filters.
	insertAddress(t, db, 1, 1, 0, 0)
	insertAddress(t, db, 2, 2, 1, 0)
	insertAddress(t, db, 3, 2, 0, 1) // deleted -> excluded

	repo := NewAddressRepository(db)

	got, err := repo.GetAddressListByUser(0, "")
	if err \!= nil {
		t.Fatalf("GetAddressListByUser returned error: %v", err)
	}

	if len(got) \!= 2 {
		t.Fatalf("expected 2 addresses, got %d", len(got))
	}

	// Order: is_default desc, id desc -> id:2 (default) then id:1
	if got[0].ID \!= 2 || got[1].ID \!= 1 {
		t.Fatalf("unexpected order, got IDs: [%d, %d], want [2, 1]", got[0].ID, got[1].ID)
	}
	if got[0].UserName \!= "" || got[1].UserName \!= "" {
		t.Fatalf("expected empty user_name when no filters; got: [%q, %q]", got[0].UserName, got[1].UserName)
	}
}

func TestGetAddressListByUser_FilterByUserID_JoinsUserAndOrders(t *testing.T) {
	db := newTestDB(t)
	insertUser(t, db, 1, "Bob")
	insertUser(t, db, 2, "Alice")

	// user 2 has one default and one non-default; one deleted row
	insertAddress(t, db, 10, 2, 0, 0)
	insertAddress(t, db, 11, 2, 1, 0)
	insertAddress(t, db, 12, 2, 0, 1)
	// user 1 has a row that should not show when filtering by userId=2
	insertAddress(t, db, 13, 1, 1, 0)

	repo := NewAddressRepository(db)
	got, err := repo.GetAddressListByUser(2, "")
	if err \!= nil {
		t.Fatalf("GetAddressListByUser returned error: %v", err)
	}

	if len(got) \!= 2 {
		t.Fatalf("expected 2 addresses for user 2, got %d", len(got))
	}
	// Order: default first (id 11), then id 10
	if got[0].ID \!= 11 || got[1].ID \!= 10 {
		t.Fatalf("unexpected order, got IDs: [%d, %d], want [11, 10]", got[0].ID, got[1].ID)
	}
	for i := range got {
		if got[i].UserName \!= "Alice" {
			t.Fatalf("expected user_name 'Alice', got %q at index %d", got[i].UserName, i)
		}
	}
}

func TestGetAddressListByUser_FilterByUserName_LikeMatch(t *testing.T) {
	db := newTestDB(t)
	insertUser(t, db, 1, "Bob")
	insertUser(t, db, 2, "Alice Wonderland")
	insertAddress(t, db, 20, 2, 0, 0)
	insertAddress(t, db, 21, 2, 1, 0)
	insertAddress(t, db, 22, 1, 1, 0) // Bob's address should not match "land"

	repo := NewAddressRepository(db)
	got, err := repo.GetAddressListByUser(0, "land")
	if err \!= nil {
		t.Fatalf("GetAddressListByUser returned error: %v", err)
	}
	if len(got) \!= 2 {
		t.Fatalf("expected 2 addresses (user 2), got %d", len(got))
	}
	// Order: default first (id 21), then id 20
	if got[0].ID \!= 21 || got[1].ID \!= 20 {
		t.Fatalf("unexpected order, got IDs: [%d, %d], want [21, 20]", got[0].ID, got[1].ID)
	}
	for _, row := range got {
		if row.UserName \!= "Alice Wonderland" {
			t.Fatalf("expected user_name 'Alice Wonderland', got %q", row.UserName)
		}
	}
}

func TestGetAddressListByUser_FilterByBoth_Intersection(t *testing.T) {
	db := newTestDB(t)
	insertUser(t, db, 1, "Charlie")
	insertUser(t, db, 2, "Chris")

	insertAddress(t, db, 30, 1, 1, 0)
	insertAddress(t, db, 31, 2, 1, 0)

	repo := NewAddressRepository(db)

	// Both filters match -> returns only user 1's addresses
	got, err := repo.GetAddressListByUser(1, "Char")
	if err \!= nil {
		t.Fatalf("GetAddressListByUser err: %v", err)
	}
	if len(got) \!= 1 || got[0].UserName \!= "Charlie" || got[0].UserID \!= 1 {
		t.Fatalf("unexpected result for intersecting filters: %+v", got)
	}

	// Mismatched filters -> empty slice, no error
	gotEmpty, err := repo.GetAddressListByUser(1, "Chris")
	if err \!= nil {
		t.Fatalf("GetAddressListByUser err on mismatch: %v", err)
	}
	if len(gotEmpty) \!= 0 {
		t.Fatalf("expected empty result for mismatched filters, got %d", len(gotEmpty))
	}
}

func TestGetDefaultOrFirstAddressID_PrefersDefaultThenFirstByID(t *testing.T) {
	db := newTestDB(t)
	repo := NewAddressRepository(db)

	// user 100: one default and one non-default
	insertAddress(t, db, 40, 100, 0, 0)
	insertAddress(t, db, 41, 100, 1, 0)

	addr, err := repo.GetDefaultOrFirstAddressID(100)
	if err \!= nil {
		t.Fatalf("expected default address, got error: %v", err)
	}
	if addr == nil || addr.ID \!= 41 {
		t.Fatalf("expected ID 41 as default, got %+v", addr)
	}

	// user 101: no default -> first by id asc among non-deleted
	insertAddress(t, db, 50, 101, 0, 0)
	insertAddress(t, db, 51, 101, 0, 0)
	first, err := repo.GetDefaultOrFirstAddressID(101)
	if err \!= nil {
		t.Fatalf("expected first non-deleted address, got error: %v", err)
	}
	if first == nil || first.ID \!= 50 {
		t.Fatalf("expected ID 50 as first by id, got %+v", first)
	}

	// user 102: only deleted -> expect not found error
	insertAddress(t, db, 60, 102, 1, 1)
	none, err := repo.GetDefaultOrFirstAddressID(102)
	if err == nil {
		t.Fatalf("expected error for only-deleted addresses, got none; result=%+v", none)
	}
}

func TestSetDefault_SetsOneDefaultAndResetsOthers(t *testing.T) {
	db := newTestDB(t)
	repo := NewAddressRepository(db)

	insertAddress(t, db, 70, 200, 0, 0)
	insertAddress(t, db, 71, 200, 1, 0)

	// Change default to id 70
	if err := repo.SetDefault(200, 70); err \!= nil {
		t.Fatalf("SetDefault returned error: %v", err)
	}

	var a, b model.Address
	if err := db.Where("id = 70").First(&a).Error; err \!= nil {
		t.Fatalf("failed to load id 70: %v", err)
	}
	if err := db.Where("id = 71").First(&b).Error; err \!= nil {
		t.Fatalf("failed to load id 71: %v", err)
	}
	if a.IsDefault \!= 1 || b.IsDefault \!= 0 {
		t.Fatalf("unexpected is_default after SetDefault: id70=%d id71=%d", a.IsDefault, b.IsDefault)
	}

	// Passing id=0 should zero-out all defaults
	if err := repo.SetDefault(200, 0); err \!= nil {
		t.Fatalf("SetDefault(0) returned error: %v", err)
	}
	if err := db.Where("id = 70").First(&a).Error; err \!= nil {
		t.Fatalf("reload id 70: %v", err)
	}
	if err := db.Where("id = 71").First(&b).Error; err \!= nil {
		t.Fatalf("reload id 71: %v", err)
	}
	if a.IsDefault \!= 0 || b.IsDefault \!= 0 {
		t.Fatalf("expected both defaults cleared; got id70=%d id71=%d", a.IsDefault, b.IsDefault)
	}
}

func TestBasicCRUD_GetById_GetByUserId_GetByUserAppointId_Delete(t *testing.T) {
	db := newTestDB(t)
	repo := NewAddressRepository(db)

	insertAddress(t, db, 80, 300, 0, 0)
	insertAddress(t, db, 81, 300, 1, 0)
	insertAddress(t, db, 82, 300, 0, 1) // deleted

	// GetById
	a, err := repo.GetById(80)
	if err \!= nil || a == nil || a.ID \!= 80 {
		t.Fatalf("GetById expected id=80; got addr=%+v err=%v", a, err)
	}

	// GetByUserId: non-deleted only; ordered by is_default desc
	as, err := repo.GetByUserId(300)
	if err \!= nil {
		t.Fatalf("GetByUserId error: %v", err)
	}
	if len(as) \!= 2 {
		t.Fatalf("expected 2 non-deleted addresses, got %d", len(as))
	}
	if as[0].IsDefault \!= 1 || as[0].ID \!= 81 || as[1].ID \!= 80 {
		t.Fatalf("unexpected order from GetByUserId: %+v", as)
	}

	// GetByUserAppointId
	a2, err := repo.GetByUserAppointId(300, 81)
	if err \!= nil || a2 == nil || a2.ID \!= 81 {
		t.Fatalf("GetByUserAppointId expected id=81; got addr=%+v err=%v", a2, err)
	}

	// Delete flips is_delete=1
	if err := repo.Delete(80); err \!= nil {
		t.Fatalf("Delete error: %v", err)
	}
	var deleted model.Address
	if err := db.Where("id = 80").First(&deleted).Error; err \!= nil {
		t.Fatalf("failed to reload deleted row: %v", err)
	}
	if deleted.IsDelete \!= 1 {
		t.Fatalf("expected is_delete=1 after Delete; got %d", deleted.IsDelete)
	}
}

// Optionally, ensure that repository methods handle unexpected inputs gracefully.
func TestGracefulHandling_OnNonexistentIDs(t *testing.T) {
	db := newTestDB(t)
	repo := NewAddressRepository(db)

	if _, err := repo.GetById(99999); err == nil {
		t.Fatalf("expected error when fetching nonexistent id")
	}
	if _, err := repo.GetByUserAppointId(12345, 67890); err == nil {
		t.Fatalf("expected error when fetching nonexistent (user,id) pair")
	}
	// SetDefault on user with no rows should still succeed (updates affect 0 rows).
	if err := repo.SetDefault(54321, 0); err \!= nil {
		t.Fatalf("SetDefault on empty user should not error, got %v", err)
	}
}

// Tiny delay helper to ensure deterministic ordering by id when inserting quickly (if needed)
func smallDelay() { time.Sleep(1 * time.Millisecond) }