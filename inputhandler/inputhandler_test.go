package inputhandler

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"testing"

	"robuku/bukudb"
	"robuku/rofi"
	"robuku/rofidata"

	_ "github.com/mattn/go-sqlite3"
)

const sqlTestDbPath string = "./bookmarks-test.db"

type mockDB struct {
	bookmarks []bukudb.Bookmark
}

func newMockDB() *mockDB {
	return &mockDB{bookmarks: []bukudb.Bookmark{
		{ID: 1, URL: "https://www.google.com", Title: "metadata (title) google",
			Tags: []string{"google", "tag2", "tag3"}, Comment: "desc (comment) google"},

		{ID: 2, URL: "https://www.b.com", Title: "metadata (title) b",
			Tags: []string{"b", "tag2", "tag3"}},

		{ID: 3, URL: "https://www.c.com", Title: "metadata (title) c"},

		{ID: 4, URL: "https://www.d.com"},
	}}
}

func (db *mockDB) Len() int {
	return len(db.bookmarks)
}

func (db *mockDB) GetAll() ([]bukudb.Bookmark, error) {
	return db.bookmarks, nil
}

func (db *mockDB) Get(id uint16) (bukudb.Bookmark, error) {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return bukudb.Bookmark{}, fmt.Errorf("id out of range")
	}
	return db.bookmarks[int(id-1)], nil
}

func (db *mockDB) Add(b bukudb.Bookmark) error {
	isNew := true
	for _, en := range db.bookmarks {
		if b.URL == en.URL {
			isNew = false
			break
		}
	}
	if isNew {
		b.ID = 1 + uint16(db.Len())
		db.bookmarks = append(db.bookmarks, b)
	}
	return nil
}

func (db *mockDB) UpdateTitle(id uint16, title string) error {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return fmt.Errorf("id out of range")
	}
	db.bookmarks[id-1].Title = title
	return nil
}

func (db *mockDB) UpdateURL(id uint16, url string) error {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return fmt.Errorf("id out of range")
	}
	db.bookmarks[id-1].URL = url
	return nil
}

func (db *mockDB) UpdateComment(id uint16, comment string) error {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return fmt.Errorf("id out of range")
	}
	db.bookmarks[id-1].Comment = comment
	return nil
}

func (db *mockDB) AddTags(id uint16, tags []string) error {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return fmt.Errorf("id out of range")
	}
	for _, t := range tags {
		if !slices.Contains(db.bookmarks[id-1].Tags, t) {
			db.bookmarks[id-1].Tags = append(db.bookmarks[id-1].Tags, t)
		}
	}
	sort.Slice(db.bookmarks[id-1].Tags, func(i, j int) bool {
		return strings.ToLower(db.bookmarks[id-1].Tags[i]) <
			strings.ToLower(db.bookmarks[id-1].Tags[j])
	})
	return nil
}

func (db *mockDB) RemoveTags(id uint16, tags []string) error {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return fmt.Errorf("id out of range")
	}
	tmp := make([]string, 0)
	for _, t := range tags {
		if !slices.Contains(db.bookmarks[id-1].Tags, t) {
			tmp = append(tmp, t)
		}
	}
	db.bookmarks[id-1].Tags = tmp
	return nil
}

func (db *mockDB) ClearTags(id uint16) error {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return fmt.Errorf("id out of range")
	}
	db.bookmarks[id-1].Tags = []string{}
	return nil
}

func (db *mockDB) Remove(id uint16) error {
	if id > uint16(len(db.bookmarks)) || id < 1 {
		return fmt.Errorf("id out of range")
	}
	db.bookmarks = slices.Delete(db.bookmarks, int(id-1), 1)
	return nil
}

func Test_HandleBookmarksShow(t *testing.T) {
	in := initInputHandler(t)
	in.HandleBookmarksShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"add: Alt+1 | modify: Alt+2 | delete: Alt+3", "", ""),
		rofi.OptionNoCustom: "true",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: "0001. metadata (title) google", Meta: "google tag2 tag3"},
		{Text: "0002. metadata (title) b", Meta: "b tag2 tag3"},
		{Text: "0003. metadata (title) c"},
		{Text: "0004. https://www.d.com"},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)

	if in.api.Data.Bookmark.ID != 0 {
		t.Errorf("expected Bookmark ID '0', got '%d'", in.api.Data.Bookmark.ID)
	}
}

func Test_handleBookmarksSelect(t *testing.T) {
	in := initInputHandler(t)

	// selected add option
	in.handleBookmarksSelect("", rofi.StateCustomKeybinding1)
	checkState(t, rofidata.St_add_select, in.api.Data.State)

	// selected modify option
	in.handleBookmarksSelect("0001. metadata (title) a", rofi.StateCustomKeybinding2)
	checkState(t, rofidata.St_modify_select, in.api.Data.State)

	// selected delete option
	in.handleBookmarksSelect("0001. metadata (title) a", rofi.StateCustomKeybinding3)
	checkState(t, rofidata.St_delete_confirm_select, in.api.Data.State)

	// selected valid bookmark
	in.handleBookmarksSelect("0001. metadata (title) a", rofi.StateSelected)
	checkState(t, rofidata.St_goto_exec, in.api.Data.State)
	if in.api.Data.Bookmark.ID != 1 {
		t.Errorf("expected Bookmark ID '1', got '%d'", in.api.Data.Bookmark.ID)
	}

	// selected invalid bookmark that has no id
	in.handleBookmarksSelect("invalid bookmark", rofi.StateSelected)
	checkState(t, rofidata.St_bookmarks_show, in.api.Data.State)

	// selected invalid bookmark that has id out of range
	in.handleBookmarksSelect("0099. invalid id", rofi.StateSelected)
	checkState(t, rofidata.St_bookmarks_show, in.api.Data.State)
}

func Test_handleAddShow(t *testing.T) {
	in := initInputHandler(t)
	in.handleAddShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"select a field to add, all are optional except the url", "", ""),
		rofi.OptionNoCustom: "true",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	b := in.api.Data.Bookmark
	b.ID = uint16(in.db.Len() + 1)
	expectedEntries := []rofi.Entry{{Text: op_back}}
	bookmark := multiLineBookmark(b)
	for _, l := range bookmark {
		expectedEntries = append(expectedEntries, rofi.Entry{Text: l})
	}
	expectedEntries = append(expectedEntries, rofi.Entry{Text: op_confirm})
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_add_select, in.api.Data.State)
}

func Test_handleAddSelect(t *testing.T) {
	in := initInputHandler(t)

	// selected back option
	in.handleAddSelect(op_back)
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)

	// selected confirm option with no url entered
	in.handleAddSelect(op_confirm)
	checkState(t, rofidata.St_add_show, in.api.Data.State)

	// selected confirm option with url entered
	in.api.Data.Bookmark.URL = "https://www.a-new-bookmark.com"
	in.handleAddSelect(op_confirm)
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)

	// selected title
	in.handleAddSelect("1. (title)")
	checkState(t, rofidata.St_add_title_select, in.api.Data.State)

	// selected url
	in.handleAddSelect("> (url)")
	checkState(t, rofidata.St_add_url_select, in.api.Data.State)

	// selected comment
	in.handleAddSelect("+ (comment)")
	checkState(t, rofidata.St_add_comment_select, in.api.Data.State)

	// selected tags
	in.handleAddSelect("# (tags)")
	checkState(t, rofidata.St_add_tags_select, in.api.Data.State)

	// selected invalid
	in.handleAddSelect("AAAAAAA")
	checkState(t, rofidata.St_add_select, in.api.Data.State)
}

func Test_handleAddTitleShow(t *testing.T) {
	in := initInputHandler(t)
	in.handleAddTitleShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage:  generatePangoMarkup("enter a title", "", ""),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_add_title_select, in.api.Data.State)
}

func Test_handleAddTitleSelect(t *testing.T) {
	in := initInputHandler(t)

	// selected back option
	in.handleAddTitleSelect(op_back)
	if in.api.Data.State != rofidata.St_add_select {
		t.Errorf("expected state '%d', got '%d'",
			rofidata.St_add_select, in.api.Data.State)
	}

	// selected default option, entered new title
	in.handleAddTitleSelect("some title")
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if in.api.Data.Bookmark.Title != "some title" {
		t.Errorf("expected bookmark title 'some title', got '%v'",
			in.api.Data.Bookmark.Title)
	}

	// selected delete option
	in.handleAddTitleSelect(op_delete)
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if in.api.Data.Bookmark.Title != "" {
		t.Errorf("expected bookmark title '', got '%v'",
			in.api.Data.Bookmark.Title)
	}
}

func Test_handleAddUrlShow(t *testing.T) {
	in := initInputHandler(t)
	in.handleAddUrlShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage:  generatePangoMarkup("enter a url", "", ""),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_add_url_select, in.api.Data.State)
}

func Test_handleAddUrlSelect(t *testing.T) {
	in := initInputHandler(t)

	// selected back option
	in.handleAddUrlSelect(op_back)
	checkState(t, rofidata.St_add_select, in.api.Data.State)

	// selected default option, entered new url
	in.handleAddUrlSelect("some url")
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if in.api.Data.Bookmark.URL != "some url" {
		t.Errorf("expected bookmark url 'some url', got '%v'",
			in.api.Data.Bookmark.URL)
	}

	// selected delete option
	in.handleAddUrlSelect(op_delete)
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if in.api.Data.Bookmark.URL != "" {
		t.Errorf("expected bookmark url '', got '%v'",
			in.api.Data.Bookmark.URL)
	}
}

func Test_handleAddCommentShow(t *testing.T) {
	in := initInputHandler(t)
	in.handleAddCommentShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage:  generatePangoMarkup("enter a comment", "", ""),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_add_comment_select, in.api.Data.State)
}

func Test_handleAddCommentSelect(t *testing.T) {
	in := initInputHandler(t)

	// selected back option
	in.handleAddCommentSelect(op_back)
	checkState(t, rofidata.St_add_select, in.api.Data.State)

	// selected default option, entered new comment
	in.handleAddCommentSelect("some comment")
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if in.api.Data.Bookmark.Comment != "some comment" {
		t.Errorf("expected bookmark comment 'some comment', got '%v'",
			in.api.Data.Bookmark.Comment)
	}

	// selected delete option
	in.handleAddCommentSelect(op_delete)
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if in.api.Data.Bookmark.Comment != "" {
		t.Errorf("expected bookmark comment '', got '%v'",
			in.api.Data.Bookmark.Comment)
	}
}

func Test_handleAddTagsShow(t *testing.T) {
	in := initInputHandler(t)
	in.handleAddTagsShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"enter some tags", "'mytag, some-tag, a tag'", ""),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_add_tags_select, in.api.Data.State)
}

func Test_handleAddTagsSelect(t *testing.T) {
	in := initInputHandler(t)

	// selected back option
	in.handleAddTagsSelect(op_back)
	checkState(t, rofidata.St_add_select, in.api.Data.State)

	// selected default option, entered new tags
	in.handleAddTagsSelect("some, tags")
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if in.api.Data.Bookmark.Tags[0] != "some" {
		t.Errorf("expected bookmark tag 'some', got '%v'",
			in.api.Data.Bookmark.Tags[0])
	}
	if in.api.Data.Bookmark.Tags[1] != "tags" {
		t.Errorf("expected bookmark tag 'tags', got '%v'",
			in.api.Data.Bookmark.Tags[1])
	}

	// selected delete option
	in.handleAddTagsSelect(op_delete)
	checkState(t, rofidata.St_add_select, in.api.Data.State)
	if len(in.api.Data.Bookmark.Tags) != 0 {
		t.Errorf("expected bookmark tags empty , got length '%v'",
			in.api.Data.Bookmark.Tags)
	}
}

func Test_handleModifyShow(t *testing.T) {
	in := initInputHandler(t)
	in.handleModifyShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage:  generatePangoMarkup("select a field to edit", "", ""),
		rofi.OptionNoCustom: "true",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{{Text: op_back}}
	bookmark := multiLineBookmark(in.api.Data.Bookmark)
	for _, l := range bookmark {
		expectedEntries = append(expectedEntries, rofi.Entry{Text: l})
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_modify_select, in.api.Data.State)
}

func Test_handleModifySelect(t *testing.T) {
	in := initInputHandler(t)

	// selected back option
	in.handleModifySelect(op_back)
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)

	// selected title
	in.handleModifySelect("1. (title)")
	checkState(t, rofidata.St_modify_title_select, in.api.Data.State)

	// selected url
	in.handleModifySelect("> (url)")
	checkState(t, rofidata.St_modify_url_select, in.api.Data.State)

	// selected comment
	in.handleModifySelect("+ (comment)")
	checkState(t, rofidata.St_modify_comment_select, in.api.Data.State)

	// selected tags
	in.handleModifySelect("# (tags)")
	checkState(t, rofidata.St_modify_tags_select, in.api.Data.State)

	// selected invalid
	in.handleModifySelect("AAAAAAA")
	checkState(t, rofidata.St_modify_select, in.api.Data.State)
}

func Test_handleModifyTitleShow(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.Title = "some title"
	in.handleModifyTitleShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"enter a new title", "", in.api.Data.Bookmark.Title),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_modify_title_select, in.api.Data.State)
}

func Test_handleModifyTitleSelect(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.ID = 1

	// selected delete option
	in.api.Data.Bookmark.Title = "some title"
	in.handleModifyTitleSelect(op_delete)
	if in.api.Data.Bookmark.Title != "" {
		t.Errorf("expected bookmark title '', got '%s'", in.api.Data.Bookmark.Title)
	}

	// selected back option
	in.handleModifyTitleSelect(op_back)
	checkState(t, rofidata.St_modify_select, in.api.Data.State)

	// entered new title
	in.handleModifyTitleSelect("some new title")
	if in.api.Data.Bookmark.Title != "some new title" {
		t.Errorf("expected bookmark title 'some new title', got '%s'", in.api.Data.Bookmark.Title)
	}
}

func Test_handleModifyUrlShow(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.URL = "some url"
	in.handleModifyUrlShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"enter a new url", "", in.api.Data.Bookmark.URL),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_modify_url_select, in.api.Data.State)
}

func Test_handleModifyUrlSelect(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.ID = 1

	// entered empty input
	in.api.Data.Bookmark.URL = "old url"
	in.handleModifyUrlSelect("")
	checkState(t, rofidata.St_modify_url_select, in.api.Data.State)
	if in.api.Data.Bookmark.URL != "old url" {
		t.Errorf("expected bookmark url 'old url', got '%s'", in.api.Data.Bookmark.URL)
	}

	// selected back option
	in.handleModifyUrlSelect(op_back)
	checkState(t, rofidata.St_modify_select, in.api.Data.State)

	// entered new url
	in.handleModifyUrlSelect("some new url")
	checkState(t, rofidata.St_modify_url_select, in.api.Data.State)
	if in.api.Data.Bookmark.URL != "some new url" {
		t.Errorf("expected bookmark url 'some new url', got '%s'", in.api.Data.Bookmark.URL)
	}
}

func Test_handleModifyCommentShow(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.Comment = "some comment"
	in.handleModifyCommentShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"enter a new comment", "", in.api.Data.Bookmark.Comment),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_modify_comment_select, in.api.Data.State)
}

func Test_handleModifyCommentSelect(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.ID = 1

	// selected delete option
	in.api.Data.Bookmark.Comment = "some comment"
	in.handleModifyCommentSelect(op_delete)
	if in.api.Data.Bookmark.Comment != "" {
		t.Errorf("expected bookmark comment '', got '%s'", in.api.Data.Bookmark.Comment)
	}

	// selected back option
	in.handleModifyCommentSelect(op_back)
	checkState(t, rofidata.St_modify_select, in.api.Data.State)

	// entered new comment
	in.handleModifyCommentSelect("some new comment")
	if in.api.Data.Bookmark.Comment != "some new comment" {
		t.Errorf("expected bookmark comment 'some new comment', got '%s'",
			in.api.Data.Bookmark.Comment)
	}
}

func Test_handleModifyTagShow(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.Tags = []string{"some tag1", "some tag2"}
	in.handleModifyTagsShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"add or remove tags",
			"'+ newtag1, ...' or '- oldtag1, ...'",
			strings.Join(in.api.Data.Bookmark.Tags, ", ")),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_modify_tags_select, in.api.Data.State)
}

func Test_handleModifyTagSelect(t *testing.T) {
	in := initInputHandler(t)
	in.api.Data.Bookmark.ID = 1

	// selected back option
	in.handleModifyTagsSelect(op_back)
	checkState(t, rofidata.St_modify_select, in.api.Data.State)

	// selected delete option
	in.api.Data.Bookmark.Tags = []string{"tag1", "tag2"}
	in.handleModifyTagsSelect(op_delete)
	if len(in.api.Data.Bookmark.Tags) != 0 {
		t.Errorf("expected bookmark tags len '0', got '%d'", len(in.api.Data.Bookmark.Tags))
	}

	// entered new tags starting with + prefix
	in.api.Data.Bookmark.Tags = []string{"tag1", "tag2"}
	in.handleModifyTagsSelect("+ wow, zow")
	if len(in.api.Data.Bookmark.Tags) != 4 {
		t.Errorf("expected bookmark tags len '%d', got '%d'",
			4, len(in.api.Data.Bookmark.Tags))
	}

	// entered existing tags starting with + prefix
	in.api.Data.Bookmark.Tags = []string{"tag1", "tag2"}
	in.handleModifyTagsSelect("+ tag1, tag2")
	if len(in.api.Data.Bookmark.Tags) != 2 {
		t.Errorf("expected bookmark tags len '%d', got '%d'",
			2, len(in.api.Data.Bookmark.Tags))
	}

	// entered existing tags starting with - prefix
	in.api.Data.Bookmark.Tags = []string{"tag1", "tag2"}
	in.handleModifyTagsSelect("- tag1, tag2")
	if len(in.api.Data.Bookmark.Tags) != 0 {
		t.Errorf("expected bookmark tags len '%d', got '%d'",
			0, len(in.api.Data.Bookmark.Tags))
	}

	// entered new tags starting with - prefix
	in.api.Data.Bookmark.Tags = []string{"tag1", "tag2"}
	in.handleModifyTagsSelect("- new1, new2")
	if len(in.api.Data.Bookmark.Tags) != 2 {
		t.Errorf("expected bookmark tags len '%d', got '%d'",
			2, len(in.api.Data.Bookmark.Tags))
	}

	// entered test without prefix, default option
	in.api.Data.Bookmark.Tags = []string{"tag1", "tag2"}
	in.handleModifyTagsSelect("AAAAAAA")
	if len(in.api.Data.Bookmark.Tags) != 2 {
		t.Errorf("expected bookmark tags len '%d', got '%d'",
			2, len(in.api.Data.Bookmark.Tags))
	}
	checkState(t, rofidata.St_modify_tags_select, in.api.Data.State)
}

func Test_handleDeleteConfirmShow(t *testing.T) {
	in := initInputHandler(t)
	in.handleDeleteConfirmShow()

	expectedOptions := map[rofi.Option]string{
		rofi.OptionMessage: generatePangoMarkup(
			"delete? (yes/No)", "", in.api.Data.Bookmark.URL),
		rofi.OptionNoCustom: "false",
	}
	checkOptions(t, expectedOptions, in.api.Options)

	expectedEntries := []rofi.Entry{
		{Text: op_back},
	}
	checkEntries(t, expectedEntries, in.api.Entries)

	checkState(t, rofidata.St_delete_confirm_select, in.api.Data.State)
}

func Test_handleDeleteConfirmSelect(t *testing.T) {
	in := initInputHandler(t)

	// selected back option
	in.api.Data.Bookmark.ID = 1
	in.handleDeleteConfirmSelect(op_back)
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)

	// did not enter 'yes'
	in.api.Data.Bookmark.ID = 1
	oldLen := in.db.Len()
	in.handleDeleteConfirmSelect("Yes")
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)
	if in.db.Len() != oldLen {
		t.Errorf("expected bookmark db len '%d', got '%d'",
			oldLen, in.db.Len())
	}
	// did not enter 'yes'
	in.api.Data.Bookmark.ID = 1
	in.handleDeleteConfirmSelect("YES")
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)
	if in.db.Len() != oldLen {
		t.Errorf("expected bookmark db len '%d', got '%d'",
			oldLen, in.db.Len())
	}
	// did not enter 'yes'
	in.api.Data.Bookmark.ID = 1
	in.handleDeleteConfirmSelect("no")
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)
	if in.db.Len() != oldLen {
		t.Errorf("expected bookmark db len '%d', got '%d'",
			oldLen, in.db.Len())
	}
	// did not enter 'yes'
	in.api.Data.Bookmark.ID = 1
	in.handleDeleteConfirmSelect("foo")
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)
	if in.db.Len() != oldLen {
		t.Errorf("expected bookmark db len '%d', got '%d'",
			oldLen, in.db.Len())
	}

	// entered 'yes'
	in.api.Data.Bookmark.ID = 1
	oldLen = in.db.Len()
	in.handleDeleteConfirmSelect("yes")
	checkState(t, rofidata.St_bookmarks_select, in.api.Data.State)
	if in.db.Len() != oldLen-1 {
		t.Errorf("expected bookmark db len '%d', got '%d'",
			oldLen-1, in.db.Len())
	}
}

func Test_getSelectedFromInput(t *testing.T) {
	in := initInputHandler(t)

	// valid input
	sel, err := in.getSelectedFromInput("0001. this is valid")
	if err != nil {
		t.Errorf("expected no error from getSelectedFromInput(), got '%v'", err)
	}
	if sel.URL != "https://www.google.com" {
		t.Errorf("expected selected bookmark to have url 'https://www.google.com', got '%v'", sel.URL)
	}

	// invalid input
	sel, err = in.getSelectedFromInput("this is invalid")
	if err == nil {
		t.Errorf("expected error from getSelectedFromInput(), got nil")
	}
	if sel.URL != "" {
		t.Errorf("expected selected bookmark to have url '', got '%v'", sel.URL)
	}
}

func checkEntries(t *testing.T, expectedEntries, actualEntries []rofi.Entry) {
	t.Helper()
	if len(actualEntries) != len(expectedEntries) {
		t.Errorf("expected Entries length '%d', got '%d'",
			len(expectedEntries), len(actualEntries))
		return
	}
	for i, eE := range expectedEntries {
		aE := actualEntries[i]
		if aE != eE {
			t.Errorf("expected Entry at index %d to be '%v', got '%v'", i, eE, aE)
		}
	}
}

func checkOptions(t *testing.T, expectedOptions, actualOptions map[rofi.Option]string) {
	t.Helper()

	for k, expected := range expectedOptions {
		actual := actualOptions[k]
		if actual != expected {
			t.Errorf("expected option '%s' to be value '%s', got '%s'",
				k, expected, actual)
		}
	}
}

func checkState(t *testing.T, expectedState, actualState rofidata.AppState) {
	t.Helper()

	if actualState != expectedState {
		t.Errorf("expected state '%d', got '%d'",
			expectedState, actualState)
	}
}

func initInputHandler(t *testing.T) *InputHandler {
	t.Helper()

	db := newMockDB()
	data := rofidata.Data{}
	api, err := rofi.NewRofiApi(&data)
	if err != nil {
		t.Fatalf("expected no error from NewRofiApi(), got %v", err)
	}
	return NewInputHandler(db, api)
}

func isMatchingBookmarkSlice(t *testing.T, expected, actual []bukudb.Bookmark) bool {
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

func isMatchingBookmark(t *testing.T, expected, actual bukudb.Bookmark) bool {
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
