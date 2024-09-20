// bukudb, for use with https://github.com/jarun/Buku
package bukudb

import (
	"database/sql"
	"fmt"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

/* buku database schema
bookmarks (
    id INTEGER PRIMARY KEY,
    URL TEXT NOT NULL UNIQUE,
    metadata TEXT DEFAULT '',
    tags TEXT DEFAULT ',',
    desc TEXT DEFAULT '',
    flags INTEGER DEFAULT 0
);
*/

// MaxBookmarks defines the maximum number of bookmarks that can be stored.
const MaxBookmarks = 1000

// Bookmark represents a single bookmark entry.
type Bookmark struct {
	// ID is the unique identifier for the bookmark.
	ID uint16

	// URL of the bookmark.
	URL string

	// Title (metadata) of the bookmark.
	Title string

	// Tags associated with the bookmark.
	Tags []string

	// Comment or description of the bookmark.
	Comment string
}

// BukuDB represents a connection to the buku SQLite database.
type BukuDB struct {
	dbPath string
	conn   *sql.DB
	mu     *sync.Mutex
	len    int
}

// NewBukuDB initializes and returns a new BukuDB instance.
func NewBukuDB(dbPath string) (*BukuDB, error) {
	mu := sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	l, err := getMaxBookmarkID(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get database length: %w", err)
	}

	return &BukuDB{
		dbPath: dbPath,
		conn:   conn,
		mu:     &mu,
		len:    l,
	}, nil
}

// Close closes the database connection.
func (db *BukuDB) Close() error {
	return db.conn.Close()
}

// Len returns the number of bookmarks in db.
func (db *BukuDB) Len() int {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.len
}

// GetAll returns a all bookmarks in db.
func (db *BukuDB) GetAll() ([]Bookmark, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return loadBookmarks(db.conn, db.len)
}

// Get returns a bookmark by ID.
func (db *BukuDB) Get(id uint16) (Bookmark, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if id < 1 || int(id) > db.len {
		return Bookmark{}, fmt.Errorf("bookmark id %d out of range (1-%d)", id, db.len)
	}

	var b Bookmark
	var tagsString string
	row := db.conn.QueryRow("SELECT id, URL, metadata, tags, desc FROM bookmarks WHERE id = ?", id)
	if err := row.Scan(&b.ID, &b.URL, &b.Title, &tagsString, &b.Comment); err != nil {
		return Bookmark{}, fmt.Errorf("failed to scan bookmark: %w", err)
	}

	if tagsString != "," {
		b.Tags = strings.Split(tagsString, ",")
		b.Tags = filter(b.Tags, func(t string) bool { return t != "" })
	}

	return b, nil
}

// Add inserts a new bookmark into the database.
func (db *BukuDB) Add(bookmark Bookmark) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	bookmark.ID = uint16(db.len + 1)
	if bookmark.ID > uint16(MaxBookmarks) {
		return fmt.Errorf("maximum number of bookmarks (%d) reached", MaxBookmarks)
	}

	query := `INSERT INTO bookmarks (id, URL, metadata, tags, desc, flags) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.conn.Exec(
		query,
		bookmark.ID,
		bookmark.URL,
		bookmark.Title,
		strings.Join(bookmark.Tags, ","),
		bookmark.Comment,
		0,
	)
	if err != nil {
		return fmt.Errorf("failed to insert bookmark: %w", err)
	}

	db.len = int(bookmark.ID)
	return nil
}

// UpdateTitle updates the title of the bookmark with the given ID.
func (db *BukuDB) UpdateTitle(id uint16, title string) error {
	return db.updateField(id, "metadata", title)
}

// UpdateURL updates the URL of the bookmark with the given ID.
func (db *BukuDB) UpdateURL(id uint16, url string) error {
	return db.updateField(id, "URL", url)
}

// UpdateComment updates the comment of the bookmark with the given ID.
func (db *BukuDB) UpdateComment(id uint16, comment string) error {
	return db.updateField(id, "desc", comment)
}

// AddTags adds tags to the bookmark with the given ID.
func (db *BukuDB) AddTags(id uint16, tags []string) error {
	b, err := db.Get(id)
	if err != nil {
		return err
	}

	tags = filter(tags, func(t string) bool { return !slices.Contains(b.Tags, t) })
	b.Tags = append(b.Tags, tags...)

	sort.Slice(b.Tags, func(i, j int) bool {
		return strings.ToLower(b.Tags[i]) < strings.ToLower(b.Tags[j])
	})

	tagsStr := "," + strings.Join(b.Tags, ",") + ","
	return db.updateField(id, "tags", tagsStr)
}

// RemoveTags removes tags from the bookmark with the given ID.
func (db *BukuDB) RemoveTags(id uint16, tags []string) error {
	b, err := db.Get(id)
	if err != nil {
		return err
	}

	b.Tags = filter(b.Tags, func(t string) bool { return !slices.Contains(tags, t) })
	tagsStr := "," + strings.Join(b.Tags, ",") + ","
	return db.updateField(id, "tags", tagsStr)
}

// ClearTags removes all tags from the bookmark with the given ID.
func (db *BukuDB) ClearTags(id uint16) error {
	return db.updateField(id, "tags", ",")
}

// Remove removes a bookmark from the database.
func (db *BukuDB) Remove(id uint16) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	index := int(id - 1)

	if id < 1 || index >= db.len {
		return fmt.Errorf("id %d out of range (1-%d)", id, db.len)
	}

	query := `DELETE FROM bookmarks WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	for i := index + 1; i < db.len; i++ {
		updateQuery := `UPDATE bookmarks SET id = ? WHERE id = ?`
		_, err := db.conn.Exec(updateQuery, i, i+1)
		if err != nil {
			return fmt.Errorf("failed to update bookmark id: %w", err)
		}
	}

	db.len -= 1
	return nil
}

// updateField updates a specific field in the database and in-memory bookmark.
func (db *BukuDB) updateField(id uint16, field, value string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if id < 1 || int(id) > db.len {
		return fmt.Errorf("id %d out of range (1-%d)", id, db.len)
	}

	query := fmt.Sprintf("UPDATE bookmarks SET %s = ? WHERE id = ?", field)
	_, err := db.conn.Exec(query, value, id)
	if err != nil {
		return fmt.Errorf("failed to update field %s: %w", field, err)
	}

	return nil
}

// Utility functions

// getMaxBookmarkID retrieves the maximum ID from the bookmarks table.
func getMaxBookmarkID(conn *sql.DB) (int, error) {
	var maxID int
	err := conn.QueryRow("SELECT MAX(id) FROM bookmarks;").Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("failed to get max ID from bookmarks: %w", err)
	}

	if maxID > MaxBookmarks {
		maxID = MaxBookmarks
	}
	return maxID, nil
}

// loadBookmarks loads all bookmarks from the database up to maxID.
func loadBookmarks(conn *sql.DB, maxID int) ([]Bookmark, error) {
	mu := sync.Mutex{}
	bookmarksMap := make(map[uint16]Bookmark)
	var wg sync.WaitGroup
	numWorkers := runtime.NumCPU()
	entriesPerWorker := (maxID + numWorkers - 1) / numWorkers

	var processErr error
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		start := i*entriesPerWorker + 1
		end := (i + 1) * entriesPerWorker
		if end > maxID {
			end = maxID
		}

		go func(start, end int) {
			defer wg.Done()
			if err := processBookmarkRange(conn, start, end, bookmarksMap, &mu); err != nil {
				processErr = fmt.Errorf("error processing bookmarks range: %w", err)
			}
		}(start, end)
	}

	wg.Wait()

	if processErr != nil {
		return []Bookmark{}, processErr
	}

	bookmarks := make([]Bookmark, 0, len(bookmarksMap))
	for _, b := range bookmarksMap {
		bookmarks = append(bookmarks, b)
	}

	sort.Slice(bookmarks, func(i, j int) bool {
		return bookmarks[i].ID < bookmarks[j].ID
	})

	return bookmarks, nil
}

// processBookmarkRange loads a range of bookmarks into the bookmarksMap.
func processBookmarkRange(conn *sql.DB, start, end int,
	bookmarksMap map[uint16]Bookmark, mu *sync.Mutex) error {
	rows, err := conn.Query("SELECT id, URL, metadata, tags, desc, flags FROM bookmarks WHERE id BETWEEN ? AND ?", start, end)
	if err != nil {
		return fmt.Errorf("failed to query bookmarks in range (%d-%d): %w", start, end, err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Bookmark
		var tagsString string
		var flags int // Ignored for now

		if err := rows.Scan(&b.ID, &b.URL, &b.Title, &tagsString, &b.Comment, &flags); err != nil {
			return fmt.Errorf("failed to scan bookmark: %w", err)
		}

		if len(tagsString) >= 2 {
			b.Tags = strings.Split(tagsString[1:len(tagsString)-1], ",")
		}

		mu.Lock()
		bookmarksMap[b.ID] = b
		mu.Unlock()
	}

	return rows.Err()
}

func filter(slice []string, predicate func(string) bool) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}
