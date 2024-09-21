// inputhandler, handles rofi input and app state
package inputhandler

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"

	"robuku/bukudb"
	"robuku/rofi"
	"robuku/rofidata"
)

const robukuBrowserEnvVar = "ROBUKU_BROWSER"
const entryMaxLen = 100

const (
	op_add       string = "--> Add"
	op_back      string = "<-- Back"
	op_confirm   string = "--> Confirm"
	op_go_to_url string = "--> Go to URL"
	op_modify    string = "--> Modify"
	op_delete    string = "--> Delete"
)

type db interface {
	Add(bukudb.Bookmark) error
	AddTags(id uint16, tags []string) error
	ClearTags(id uint16) error
	Get(id uint16) (bukudb.Bookmark, error)
	GetAll() ([]bukudb.Bookmark, error)
	Len() int
	Remove(id uint16) error
	RemoveTags(id uint16, tags []string) error
	UpdateComment(id uint16, comment string) error
	UpdateTitle(id uint16, title string) error
	UpdateURL(id uint16, url string) error
}

// InputHandler is the struct that handles input from rofi and manages app state
type InputHandler struct {
	db      db
	api     *rofi.RofiApi[*rofidata.Data]
	data    *rofidata.Data
	browser string
}

// NewInputHandler returns a new instance of the InputHandler struct
func NewInputHandler(db db, api *rofi.RofiApi[*rofidata.Data]) *InputHandler {
	in := InputHandler{
		db:      db,
		api:     api,
		browser: os.Getenv(robukuBrowserEnvVar),
	}
	return &in
}

// HandleInput takes the selected rofi entry/input and processes it based on app state
func (in *InputHandler) HandleInput(input string) {
	input = strings.TrimSpace(input)
	rofiState := in.api.GetState()

	switch in.api.Data.State {
	case rofidata.St_bookmarks_show:
		in.HandleBookmarksShow()
	case rofidata.St_bookmarks_select:
		in.handleBookmarksSelect(input, rofiState)
	case rofidata.St_add_show:
		in.handleAddShow()
	case rofidata.St_add_select:
		in.handleAddSelect(input)
	case rofidata.St_add_title_show:
		in.handleAddTitleShow()
	case rofidata.St_add_title_select:
		in.handleAddTitleSelect(input)
	case rofidata.St_add_url_show:
		in.handleAddUrlShow()
	case rofidata.St_add_url_select:
		in.handleAddUrlSelect(input)
	case rofidata.St_add_comment_show:
		in.handleAddCommentShow()
	case rofidata.St_add_comment_select:
		in.handleAddCommentSelect(input)
	case rofidata.St_add_tags_show:
		in.handleAddTagsShow()
	case rofidata.St_add_tags_select:
		in.handleAddTagsSelect(input)
	case rofidata.St_modify_show:
		in.handleModifyShow()
	case rofidata.St_modify_select:
		in.handleModifySelect(input)
	case rofidata.St_modify_title_show:
		in.handleModifyTitleShow()
	case rofidata.St_modify_title_select:
		in.handleModifyTitleSelect(input)
	case rofidata.St_modify_url_show:
		in.handleModifyUrlShow()
	case rofidata.St_modify_url_select:
		in.handleModifyUrlSelect(input)
	case rofidata.St_modify_comment_show:
		in.handleModifyCommentShow()
	case rofidata.St_modify_comment_select:
		in.handleModifyCommentSelect(input)
	case rofidata.St_modify_tags_show:
		in.handleModifyTagsShow()
	case rofidata.St_modify_tags_select:
		in.handleModifyTagsSelect(input)
	case rofidata.St_delete_confirm_show:
		in.handleDeleteConfirmShow()
	case rofidata.St_delete_confirm_select:
		in.handleDeleteConfirmSelect(input)
	default:
		log.Printf("Unhandled state: %v", in.api.Data.State)
	}
}

// HandleBookmarksShow sets rofi's initial state and shows all bookmarks
func (in *InputHandler) HandleBookmarksShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"add: Alt+1 | modify: Alt+2 | delete: Alt+3", "", "")
	in.api.Options[rofi.OptionNoCustom] = "true"
	in.api.Options[rofi.OptionUseHotKeys] = "true"

	numPadding := len(fmt.Sprint(bukudb.MaxBookmarks))
	allBookmarks, err := in.db.GetAll()
	if err != nil {
		SetMessageToError(in.api, err)
		in.api.Data.State = rofidata.St_bookmarks_show
		return
	}
	entries := make([]rofi.Entry, 0, in.db.Len())
	for _, b := range allBookmarks {
		id := fmt.Sprint(b.ID)
		for j := len(id); j < numPadding; j++ {
			id = "0" + id
		}

		text := b.Title
		if b.Title == "" {
			text = b.URL
		}

		entries = append(entries, rofi.Entry{
			Text: formatEntryText(fmt.Sprintf("%s. %s", id, text)),
			Meta: strings.Join(b.Tags, " "),
		})
	}

	in.api.Entries = entries
	in.api.Data.State = rofidata.St_bookmarks_select
	in.api.Data.Bookmark = bukudb.Bookmark{}
}

func (in *InputHandler) handleBookmarksSelect(input string, rofiState rofi.State) {
	if rofiState == rofi.StateCustomKeybinding1 {
		in.handleAddShow()
		return
	}

	id, err := getIdFromBookmarkString(input)
	if err != nil {
		SetMessageToError(in.api, err)
		in.api.Data.State = rofidata.St_bookmarks_show
		return
	}

	b, err := in.db.Get(id)
	if err != nil {
		SetMessageToError(in.api, err)
		in.api.Data.State = rofidata.St_bookmarks_show
		return
	}

	in.api.Data.Bookmark = b

	switch rofiState {
	case rofi.StateCustomKeybinding2:
		in.handleModifyShow()
	case rofi.StateCustomKeybinding3:
		in.handleDeleteConfirmShow()
	case rofi.StateSelected:
		in.handleGotoExec()
	default:
		in.HandleBookmarksShow()
	}
}

func (in *InputHandler) handleAddShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"select a field to add, all are optional except the url", "", "")
	in.api.Options[rofi.OptionNoCustom] = "true"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	b := in.api.Data.Bookmark
	b.ID = uint16(in.db.Len() + 1)
	entries := []rofi.Entry{{Text: op_back}}
	bookmark := multiLineBookmark(b)
	for _, l := range bookmark {
		entries = append(entries, rofi.Entry{Text: l})
	}
	entries = append(entries, rofi.Entry{Text: op_confirm})
	in.api.Entries = entries

	in.api.Data.State = rofidata.St_add_select
}

func (in *InputHandler) handleAddSelect(input string) {
	if input == op_back {
		in.HandleBookmarksShow()
		return
	}

	if input == op_confirm {
		if in.api.Data.Bookmark.URL == "" {
			SetMessageToError(in.api, fmt.Errorf("error: bookmark has no url"))
			in.api.Data.State = rofidata.St_add_show
			return
		}
		err := in.db.Add(in.api.Data.Bookmark)
		if err != nil {
			SetMessageToError(in.api, err)
			in.api.Data.State = rofidata.St_bookmarks_show
			return
		}
		in.HandleBookmarksShow()
		return
	}

	switch input[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		in.handleAddTitleShow()
	case '>':
		in.handleAddUrlShow()
	case '+':
		in.handleAddCommentShow()
	case '#':
		in.handleAddTagsShow()
	default:
		in.handleAddShow()
	}
}

func (in *InputHandler) handleAddTitleShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup("enter a title", "", "")
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}

	in.api.Data.State = rofidata.St_add_title_select
}

func (in *InputHandler) handleAddTitleSelect(input string) {
	switch input {
	case op_back:
		break
	case op_delete:
		in.api.Data.Bookmark.Title = ""
	default:
		in.api.Data.Bookmark.Title = input
	}
	in.handleAddShow()
}

func (in *InputHandler) handleAddUrlShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup("enter a url", "", "")
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}

	in.api.Data.State = rofidata.St_add_url_select
}

func (in *InputHandler) handleAddUrlSelect(input string) {
	switch input {
	case op_back:
		break
	case op_delete:
		in.api.Data.Bookmark.URL = ""
	default:
		in.api.Data.Bookmark.URL = input
	}
	in.handleAddShow()
}

func (in *InputHandler) handleAddCommentShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup("enter a comment", "", "")
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}

	in.api.Data.State = rofidata.St_add_comment_select
}

func (in *InputHandler) handleAddCommentSelect(input string) {
	switch input {
	case op_back:
		break
	case op_delete:
		in.api.Data.Bookmark.Comment = ""
	default:
		in.api.Data.Bookmark.Comment = input
	}
	in.handleAddShow()
}

func (in *InputHandler) handleAddTagsShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"enter some tags", "'mytag, some-tag, a tag'", "")
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}

	in.api.Data.State = rofidata.St_add_tags_select
}

func (in *InputHandler) handleAddTagsSelect(input string) {
	switch input {
	case op_back:
		break
	case op_delete:
		in.api.Data.Bookmark.Tags = []string{}
	default:
		tags := strings.Split(input, ",")
		for i, t := range tags {
			tags[i] = strings.TrimSpace(t)
		}
		in.api.Data.Bookmark.Tags = tags

		sort.Slice(in.api.Data.Bookmark.Tags, func(i, j int) bool {
			return strings.ToLower(in.api.Data.Bookmark.Tags[i]) <
				strings.ToLower(in.api.Data.Bookmark.Tags[j])
		})
	}
	in.handleAddShow()
}

func (in *InputHandler) handleGotoExec() {
	in.api.Data.State = rofidata.St_goto_exec
	if in.browser != "" {
		cmd := exec.Command(in.browser, in.api.Data.Bookmark.URL)
		if err := cmd.Start(); err != nil {
			SetMessageToError(in.api, fmt.Errorf("error opening URL: %w", err))
			in.api.Data.State = rofidata.St_bookmarks_show
		}
	} else {
		cmd := exec.Command("xdg-open", in.api.Data.Bookmark.URL)
		if err := cmd.Start(); err != nil {
			SetMessageToError(in.api, fmt.Errorf(
				"error opening URL: xdg-utils is not installed, to use without set env variable $%s",
				robukuBrowserEnvVar))
			in.api.Data.State = rofidata.St_bookmarks_show
		}
	}
}

func (in *InputHandler) handleModifyShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"select a field to edit", "", "")
	in.api.Options[rofi.OptionNoCustom] = "true"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	entries := []rofi.Entry{{Text: op_back}}
	bookmark := multiLineBookmark(in.api.Data.Bookmark)
	for _, l := range bookmark {
		entries = append(entries, rofi.Entry{Text: l})
	}

	in.api.Entries = entries
	in.api.Data.State = rofidata.St_modify_select
}

func (in *InputHandler) handleModifySelect(input string) {
	if len(input) < 1 {
		in.handleModifyShow()
		return
	}

	if input == op_back {
		in.HandleBookmarksShow()
		return
	}

	switch input[0] {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		in.handleModifyTitleShow()
	case '>':
		in.handleModifyUrlShow()
	case '+':
		in.handleModifyCommentShow()
	case '#':
		in.handleModifyTagsShow()
	default:
		in.handleModifyShow()
	}
}

func (in *InputHandler) handleModifyTitleShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"enter a new title", "", in.api.Data.Bookmark.Title)
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}

	in.api.Data.State = rofidata.St_modify_title_select
}

func (in *InputHandler) handleModifyTitleSelect(input string) {
	if input == op_delete {
		input = ""
	}

	if input == op_back {
		in.handleModifyShow()
	} else if err := in.db.UpdateTitle(in.api.Data.Bookmark.ID, input); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error updating title: %w", err))
		in.api.Data.State = rofidata.St_modify_title_select
	} else {
		in.api.Data.Bookmark.Title = input
		in.handleModifyTitleShow()
	}
}

func (in *InputHandler) handleModifyUrlShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"enter a new url", "", in.api.Data.Bookmark.URL)
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
	}

	in.api.Data.State = rofidata.St_modify_url_select
}

func (in *InputHandler) handleModifyUrlSelect(input string) {
	if input == "" {
		in.handleModifyUrlShow()
	} else if input == op_back {
		in.handleModifyShow()
	} else if err := in.db.UpdateURL(in.api.Data.Bookmark.ID, input); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error updating url: %w", err))
		in.api.Data.State = rofidata.St_modify_url_select
	} else {
		in.api.Data.Bookmark.URL = input
		in.handleModifyUrlShow()
	}
}

func (in *InputHandler) handleModifyCommentShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"enter a new comment", "", in.api.Data.Bookmark.Comment)
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}

	in.api.Data.State = rofidata.St_modify_comment_select
}

func (in *InputHandler) handleModifyCommentSelect(input string) {
	if input == op_delete {
		input = ""
	}

	if input == op_back {
		in.handleModifyShow()
	} else if err := in.db.UpdateComment(in.api.Data.Bookmark.ID, input); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error updating comment: %w", err))
		in.api.Data.State = rofidata.St_modify_comment_select
	} else {
		in.api.Data.Bookmark.Comment = input
		in.handleModifyCommentShow()
	}
}

func (in *InputHandler) handleModifyTagsShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"add or remove tags",
		"'+ newtag1, ...' or '- oldtag1, ...'",
		strings.Join(in.api.Data.Bookmark.Tags, ", "))
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
		{Text: op_delete},
	}

	in.api.Data.State = rofidata.St_modify_tags_select
}

func (in *InputHandler) handleModifyTagsSelect(input string) {
	switch {
	case input == op_back:
		in.handleModifyShow()
	case input == op_delete:
		if err := in.db.ClearTags(in.api.Data.Bookmark.ID); err != nil {
			SetMessageToError(in.api, fmt.Errorf("error clearing tags: %w", err))
			in.api.Data.State = rofidata.St_modify_tags_select
		} else {
			in.api.Data.Bookmark.Tags = []string{}
			in.handleModifyTagsShow()
		}
	case strings.HasPrefix(input, "+"):
		tags := getTagsFromInput(input[1:])
		if err := in.db.AddTags(in.api.Data.Bookmark.ID, tags); err != nil {
			SetMessageToError(in.api, fmt.Errorf("error adding tag: %w", err))
			in.api.Data.State = rofidata.St_modify_tags_select
		} else {
			for _, t := range tags {
				if !slices.Contains(in.api.Data.Bookmark.Tags, t) {
					in.api.Data.Bookmark.Tags = append(in.api.Data.Bookmark.Tags, t)
				}
			}

			sort.Slice(in.api.Data.Bookmark.Tags, func(i, j int) bool {
				return strings.ToLower(in.api.Data.Bookmark.Tags[i]) <
					strings.ToLower(in.api.Data.Bookmark.Tags[j])
			})
			in.handleModifyTagsShow()
		}
	case strings.HasPrefix(input, "-"):
		tags := getTagsFromInput(input[1:])
		if err := in.db.RemoveTags(in.api.Data.Bookmark.ID, tags); err != nil {
			SetMessageToError(in.api, fmt.Errorf("error removing tag: %w", err))
			in.api.Data.State = rofidata.St_modify_tags_select
		} else {
			tmp := make([]string, 0)
			for _, t := range in.api.Data.Bookmark.Tags {
				if !slices.Contains(tags, t) {
					tmp = append(tmp, t)
				}
			}
			in.api.Data.Bookmark.Tags = tmp
			in.handleModifyTagsShow()
		}
	default:
		in.handleModifyTagsShow()
	}
}

func (in *InputHandler) handleDeleteConfirmShow() {
	in.api.Options[rofi.OptionMessage] = generatePangoMarkup(
		"delete? (yes/No)", "", in.api.Data.Bookmark.URL)
	in.api.Options[rofi.OptionNoCustom] = "false"
	in.api.Options[rofi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofi.Entry{
		{Text: op_back},
	}

	in.api.Data.State = rofidata.St_delete_confirm_select
}

func (in *InputHandler) handleDeleteConfirmSelect(input string) {
	if input == op_back || input != "yes" {
		in.HandleBookmarksShow()
		return
	}

	if err := in.db.Remove(in.api.Data.Bookmark.ID); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error deleting bookmark: %w", err))
		in.api.Data.State = rofidata.St_bookmarks_show
	} else {
		in.HandleBookmarksShow()
	}
}

func (in *InputHandler) getSelectedFromInput(input string) (bukudb.Bookmark, error) {
	id, err := getIdFromBookmarkString(input)
	if err != nil {
		return bukudb.Bookmark{}, err
	}
	b, _ := in.db.Get(id)
	return b, nil
}

// SetMessageToError sets rofi's message box to the text of an error and
// replaces rofi's entries with the back option
func SetMessageToError(api *rofi.RofiApi[*rofidata.Data], err error) {
	log.Println(err)
	api.Options[rofi.OptionMessage] = fmt.Sprintf(
		"<markup><span font_weight=\"bold\">error:</span><span> %s</span></markup>",
		escapePangoMarkup(err.Error()))
	api.Options[rofi.OptionNoCustom] = "true"
	api.Entries = []rofi.Entry{{Text: op_back}}
}

func getIdFromBookmarkString(input string) (uint16, error) {
	idString := strings.Split(input, ".")[0]
	idUint64, err := strconv.ParseUint(idString, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("error parsing id from entry: %s", input)
	}
	return uint16(idUint64), nil
}

func getTagsFromInput(input string) []string {
	tags := strings.Split(input, ",")
	for i, t := range tags {
		tags[i] = strings.TrimSpace(t)
	}
	return tags
}

func multiLineBookmark(b bukudb.Bookmark) []string {
	title := b.Title
	if title == "" {
		title = "(Title)"
	}

	url := b.URL
	if url == "" {
		url = "(Url)"
	}

	comment := b.Comment
	if comment == "" {
		comment = "(Comment)"
	}

	tags := strings.Join(b.Tags, ", ")
	if tags == "" {
		tags = "(Tags)"
	}

	return []string{
		formatEntryText(fmt.Sprintf("%d. %s", b.ID, title)),
		formatEntryText("   > " + url),
		formatEntryText("   + " + comment),
		formatEntryText("   # " + tags),
	}
}

func generatePangoMarkup(instructions, example, currentValue string) string {
	markup := "<markup>"

	if instructions != "" {
		instructions = escapePangoMarkup(instructions)
		markup += fmt.Sprintf(
			"<span font_weight=\"bold\">%s</span>\r", instructions)
	}
	if example != "" {
		example = escapePangoMarkup(example)
		markup += fmt.Sprintf(
			"<span font_weight=\"bold\">example:</span><span> <i>%s</i></span>\r",
			example)
	}
	if currentValue != "" {
		currentValue = truncate(currentValue, entryMaxLen)
		currentValue = escapePangoMarkup(currentValue)
		markup += fmt.Sprintf(
			"<span font_weight=\"bold\">current:</span><span> <u>%s</u></span>",
			currentValue)
	}

	markup += "</markup>"
	return markup
}

func formatEntryText(e string) string {
	e = truncate(e, entryMaxLen)
	e = replaceNewlines(e)
	return e
}

func escapePangoMarkup(input string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"'", "&#39;",
		"\"", "&quot;",
		"\n", "\r",
	)
	return replacer.Replace(input)
}

func truncate(s string, l int) string {
	if len(s) > l && l >= 3 {
		return s[0:l-3] + "..."
	} else {
		return s
	}
}

func replaceNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}
