package bukudb

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const sqlTestDbPath string = "./bookmarks-test.db"

func Test_GetAll(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	var expectedBookmarks = []Bookmark{
		{ID: 1, URL: "https://www.a.com", Title: "metadata (title) a",
			Tags: []string{"a", "tag2", "tag3"}, Comment: "desc (comment) a"},

		{ID: 2, URL: "https://www.b.com", Title: "metadata (title) b",
			Tags: []string{"b", "tag2", "tag3"}},

		{ID: 3, URL: "https://www.c.com", Title: "metadata (title) c"},

		{ID: 4, URL: "https://www.d.com"},
	}

	bs, err := db.GetAll()
	if err != nil {
		t.Fatalf("expected no error on GetAll(), got '%v'", err)
	}

	if !isMatchingBookmarkSlice(t, expectedBookmarks, bs) {
		t.Fatal("bookmarks slice does not match expected")
	}
}

func Test_Get(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 2, URL: "https://www.b.com", Title: "metadata (title) b",
		Tags: []string{"b", "tag2", "tag3"}}

	actual, err := db.Get(2)
	if err != nil {
		t.Fatalf("expected ID 2 to cause no err, got %v", err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}

	_, err = db.Get(10)
	if err == nil {
		t.Fatal("expected ID 10 to cause err, got nil")
	}
}

func Test_Add_And_Remove(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 5, URL: "https://www.new.com", Title: "metadata (title) new",
		Tags: []string{"new", "tag2", "tag3"}}

	oldLen := db.Len()

	err = db.Add(expected)
	if err != nil {
		t.Fatalf("expected no error on Add(), got '%v'", err)
	}

	if oldLen+1 != db.Len() {
		t.Fatalf("expected bookmarks length = %d, got %d", oldLen+1, db.Len())
	}

	actual, err := db.Get(expected.ID)
	if err != nil {
		t.Fatalf("expected ID '%d' to cause no err, got %v", expected.ID, err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}

	oldLen = db.Len()

	err = db.Remove(expected.ID)
	if err != nil {
		t.Fatalf("expected no error on Remove(), got '%v'", err)
	}

	if oldLen-1 != db.Len() {
		t.Fatalf("expected bookmarks length = %d, got %d", oldLen-1, db.Len())
	}
}

func Test_UpdateTitle(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 2, URL: "https://www.b.com", Title: "metadata (title) new title",
		Tags: []string{"b", "tag2", "tag3"}, Comment: ""}

	err = db.UpdateTitle(expected.ID, expected.Title)
	if err != nil {
		t.Fatalf("expected no error on UpdateTitle(), got '%v'", err)
	}

	actual, err := db.Get(expected.ID)
	if err != nil {
		t.Fatalf("expected ID '%d' to cause no err, got %v", expected.ID, err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}
}

func Test_UpdateURL(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 2, URL: "https://www.new.com", Title: "metadata (title) b",
		Tags: []string{"b", "tag2", "tag3"}, Comment: ""}

	err = db.UpdateURL(expected.ID, expected.URL)
	if err != nil {
		t.Fatalf("expected no error on UpdateURL(), got '%v'", err)
	}

	actual, err := db.Get(expected.ID)
	if err != nil {
		t.Fatalf("expected ID '%d' to cause no err, got %v", expected.ID, err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}
}

func Test_UpdateComment(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 2, URL: "https://www.b.com", Title: "metadata (title) b",
		Tags: []string{"b", "tag2", "tag3"}, Comment: "new comment"}

	err = db.UpdateComment(expected.ID, expected.Comment)
	if err != nil {
		t.Fatalf("expected no error on UpdateComment(), got '%v'", err)
	}

	actual, err := db.Get(expected.ID)
	if err != nil {
		t.Fatalf("expected ID '%d' to cause no err, got %v", expected.ID, err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}
}

func Test_AddTags(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 2, URL: "https://www.b.com", Title: "metadata (title) b",
		Tags: []string{"b", "tag2", "tag3", "tag4", "tag5"}}

	err = db.AddTags(expected.ID, expected.Tags)
	if err != nil {
		t.Fatalf("expected no error on AddTags(), got '%v'", err)
	}

	actual, err := db.Get(expected.ID)
	if err != nil {
		t.Fatalf("expected ID '%d' to cause no err, got %v", expected.ID, err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}
}

func Test_RemoveTags(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 2, URL: "https://www.b.com", Title: "metadata (title) b",
		Tags: []string{"b"}}

	err = db.RemoveTags(expected.ID, []string{"tag2", "tag3"})
	if err != nil {
		t.Fatalf("expected no error on RemoveTags(), got '%v'", err)
	}

	actual, err := db.Get(expected.ID)
	if err != nil {
		t.Fatalf("expected ID '%d' to cause no err, got %v", expected.ID, err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}
}

func Test_ClearTags(t *testing.T) {
	createTestDb(t)
	db, err := NewBukuDB(sqlTestDbPath)
	defer cleanUpTestDB(t, db)

	if err != nil {
		t.Fatalf("expected no error on NewBukuDB(), got '%v'", err)
	}

	expected := Bookmark{ID: 2, URL: "https://www.b.com", Title: "metadata (title) b",
		Tags: []string{}}

	err = db.ClearTags(expected.ID)
	if err != nil {
		t.Fatalf("expected no error on ClearTags(), got '%v'", err)
	}

	actual, err := db.Get(expected.ID)
	if err != nil {
		t.Fatalf("expected ID '%d' to cause no err, got %v", expected.ID, err)
	}

	if !isMatchingBookmark(t, expected, actual) {
		t.Fatalf("expected bookmark '%v', got '%v'", expected, actual)
	}
}

func createTestDb(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(sqlTestDbPath); err == nil {
		if err := os.Remove(sqlTestDbPath); err != nil {
			t.Fatal(err)
		}
	}

	db, err := sql.Open("sqlite3", sqlTestDbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
    CREATE TABLE IF NOT EXISTS bookmarks (
        id INTEGER PRIMARY KEY,
        URL TEXT NOT NULL UNIQUE,
        metadata TEXT DEFAULT '',
        tags TEXT DEFAULT ',',
        desc TEXT DEFAULT '',
        flags INTEGER DEFAULT 0
    );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		t.Fatalf("%q: %s\n", err, sqlStmt)
	}

	type sqlEntry struct {
		id       int
		url      string
		metadata string
		tags     string
		desc     string
		flag     int
	}

	var testSqlEntries = []sqlEntry{
		{1, "https://www.a.com", "metadata (title) a", ",a,tag2,tag3,", "desc (comment) a", 0},
		{2, "https://www.b.com", "metadata (title) b", ",b,tag2,tag3,", "", 0},
		{3, "https://www.c.com", "metadata (title) c", ",", "", 0},
		{4, "https://www.d.com", "", ",", "", 0},
	}

	query := "INSERT INTO bookmarks (id, URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?, ?)"
	for _, en := range testSqlEntries {
		_, err = db.Exec(
			query,
			en.id,
			en.url,
			en.metadata,
			en.tags,
			en.desc,
			en.flag,
		)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func cleanUpTestDB(t *testing.T, db *BukuDB) {
	t.Helper()
	if _, err := os.Stat(sqlTestDbPath); err == nil {
		db.Close()
		if err := os.Remove(sqlTestDbPath); err != nil {
			t.Fatal(err)
		}
	}
}

func isMatchingBookmarkSlice(t *testing.T, expected, actual []Bookmark) bool {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("expected bookmarks length '%d', got '%d'",
			len(expected), len(actual))
		return false
	}

	match := true
	for i := 0; i < len(expected); i++ {
		ok := isMatchingBookmark(t, expected[i], actual[i])
		if !ok && match {
			match = false
		}
	}

	return match
}

func isMatchingBookmark(t *testing.T, expected, actual Bookmark) bool {
	t.Helper()

	match := true

	if expected.ID != actual.ID {
		t.Errorf("expected bookmark ID '%d', got '%d'",
			expected.ID, actual.ID)
		match = false
	}

	if expected.URL != actual.URL {
		t.Errorf("expected bookmark URL '%s', got '%s'",
			expected.URL, actual.URL)
		match = false
	}

	if expected.Title != actual.Title {
		t.Errorf("expected bookmark Title '%s', got '%s'",
			expected.Title, actual.Title)
		match = false
	}

	if len(expected.Tags) != len(actual.Tags) {
		t.Errorf("expected bookmark Tags length '%d', got '%d'",
			len(expected.Tags), len(actual.Tags))
		match = false
	} else {
		for j := 0; j < len(expected.Tags); j++ {
			if expected.Tags[j] != actual.Tags[j] {
				t.Errorf("expected bookmark Tag '%s', got '%s'",
					expected.Tags[j], actual.Tags[j])
				match = false
			}
		}
	}

	return match
}
