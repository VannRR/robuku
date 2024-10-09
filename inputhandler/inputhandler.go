// inputhandler, handles rofi input and app state
package inputhandler

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/VannRR/robuku/bukudb"
	rofiapi "github.com/VannRR/rofi-api"
)

const robukuBrowserEnvVar = "ROBUKU_BROWSER"
const entryMaxLen = 100

type State byte

const (
	StateNull                State = iota // 0
	StateErrorShow                        // 1
	StateErrorSelect                      // 2
	StateBookmarksShow                    // 3
	StateBookmarksSelect                  // 4
	StateAddShow                          // 5
	StateAddSelect                        // 6
	StateAddTitleShow                     // 7
	StateAddTitleSelect                   // 8
	StateAddUrlShow                       // 9
	StateAddUrlSelect                     // 10
	StateAddCommentShow                   // 11
	StateAddCommentSelect                 // 12
	StateAddTagsShow                      // 13
	StateAddTagsSelect                    // 14
	StateGotoExec                         // 15
	StateModifyShow                       // 16
	StateModifySelect                     // 17
	StateModifyTitleShow                  // 18
	StateModifyTitleSelect                // 19
	StateModifyUrlShow                    // 20
	StateModifyUrlSelect                  // 21
	StateModifyCommentShow                // 22
	StateModifyCommentSelect              // 23
	StateModifyTagsShow                   // 24
	StateModifyTagsSelect                 // 25
	StateDeleteConfirmShow                // 26
	StateDeleteConfirmSelect              // 27
)

const (
	opAdd     string = "--> Add"
	opExit    string = "--> Exit"
	opBack    string = "<-- Back"
	opConfirm string = "--> Confirm"
	opModify  string = "--> Modify"
	opDelete  string = "--> Delete"
)

type Data struct {
	Bookmark bukudb.Bookmark
	State    State
}

// InputHandler is the struct that handles input from rofi and manages app state
type InputHandler struct {
	db      bukudb.DBInterface
	api     *rofiapi.RofiApi[Data]
	browser string
}

// NewInputHandler returns a new instance of the InputHandler struct
func NewInputHandler(db bukudb.DBInterface, api *rofiapi.RofiApi[Data]) *InputHandler {
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
	case StateBookmarksShow:
		in.HandleBookmarksShow()
	case StateBookmarksSelect:
		in.handleBookmarksSelect(input, rofiState)
	case StateAddShow:
		in.handleAddShow()
	case StateAddSelect:
		in.handleAddSelect(input)
	case StateAddTitleShow:
		in.handleAddTitleShow()
	case StateAddTitleSelect:
		in.handleAddTitleSelect(input)
	case StateAddUrlShow:
		in.handleAddUrlShow()
	case StateAddUrlSelect:
		in.handleAddUrlSelect(input)
	case StateAddCommentShow:
		in.handleAddCommentShow()
	case StateAddCommentSelect:
		in.handleAddCommentSelect(input)
	case StateAddTagsShow:
		in.handleAddTagsShow()
	case StateAddTagsSelect:
		in.handleAddTagsSelect(input)
	case StateModifyShow:
		in.handleModifyShow()
	case StateModifySelect:
		in.handleModifySelect(input)
	case StateModifyTitleShow:
		in.handleModifyTitleShow()
	case StateModifyTitleSelect:
		in.handleModifyTitleSelect(input)
	case StateModifyUrlShow:
		in.handleModifyUrlShow()
	case StateModifyUrlSelect:
		in.handleModifyUrlSelect(input)
	case StateModifyCommentShow:
		in.handleModifyCommentShow()
	case StateModifyCommentSelect:
		in.handleModifyCommentSelect(input)
	case StateModifyTagsShow:
		in.handleModifyTagsShow()
	case StateModifyTagsSelect:
		in.handleModifyTagsSelect(input)
	case StateDeleteConfirmShow:
		in.handleDeleteConfirmShow()
	case StateDeleteConfirmSelect:
		in.handleDeleteConfirmSelect(input)
	default:
		log.Printf("Unhandled state: %v", in.api.Data.State)
	}
}

// HandleBookmarksShow sets rofi's initial state and shows all bookmarks
func (in *InputHandler) HandleBookmarksShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"add: Alt+1 | modify: Alt+2 | delete: Alt+3", "", "")
	in.api.Options[rofiapi.OptionNoCustom] = "true"
	in.api.Options[rofiapi.OptionUseHotKeys] = "true"

	numPadding := len(fmt.Sprint(bukudb.MaxBookmarks))
	allBookmarks, err := in.db.GetAll()
	if err != nil {
		SetMessageToError(in.api, err)
		return
	}
	entries := make([]rofiapi.Entry, 0, in.db.Len())
	for _, b := range allBookmarks {
		id := fmt.Sprint(b.ID)
		for j := len(id); j < numPadding; j++ {
			id = "0" + id
		}

		text := b.Title
		meta := strings.Join(b.Tags, " ")

		if b.Title == "" {
			text = b.URL
		} else {
			if meta != "" {
				meta += " " + cleanURL(b.URL)
			} else {
				meta = cleanURL(b.URL)
			}
		}

		entries = append(entries, rofiapi.Entry{
			Text: formatEntryText(fmt.Sprintf("%s. %s", id, text)),
			Meta: meta,
		})
	}

	in.api.Entries = entries
	in.api.Data.State = StateBookmarksSelect
	in.api.Data.Bookmark = bukudb.Bookmark{}
}

func (in *InputHandler) handleBookmarksSelect(input string, rofiState rofiapi.State) {
	if rofiState == rofiapi.StateCustomKeybinding1 {
		in.handleAddShow()
		return
	}

	id, err := getIdFromBookmarkString(input)
	if err != nil {
		SetMessageToError(in.api, err)
		return
	}

	b, err := in.db.Get(id)
	if err != nil {
		SetMessageToError(in.api, err)
		return
	}

	in.api.Data.Bookmark = b

	switch rofiState {
	case rofiapi.StateCustomKeybinding2:
		in.handleModifyShow()
	case rofiapi.StateCustomKeybinding3:
		in.handleDeleteConfirmShow()
	case rofiapi.StateSelected:
		in.handleGotoExec()
	default:
		in.HandleBookmarksShow()
	}
}

func (in *InputHandler) handleAddShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"select a field to add, all are optional except the url", "", "")
	in.api.Options[rofiapi.OptionNoCustom] = "true"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	b := in.api.Data.Bookmark
	b.ID = uint16(in.db.Len() + 1)
	entries := []rofiapi.Entry{{Text: opBack}}
	bookmark := multiLineBookmark(b)
	for _, l := range bookmark {
		entries = append(entries, rofiapi.Entry{Text: l})
	}
	entries = append(entries, rofiapi.Entry{Text: opConfirm})
	in.api.Entries = entries

	in.api.Data.State = StateAddSelect
}

func (in *InputHandler) handleAddSelect(input string) {
	if input == opBack {
		in.HandleBookmarksShow()
		return
	}

	if input == opConfirm {
		if in.api.Data.Bookmark.URL == "" {
			SetMessageToError(in.api, fmt.Errorf("error: bookmark has no url"))
			return
		}
		err := in.db.Add(in.api.Data.Bookmark)
		if err != nil {
			SetMessageToError(in.api, err)
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
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup("enter a title", "", "")
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
		{Text: opDelete},
	}

	in.api.Data.State = StateAddTitleSelect
}

func (in *InputHandler) handleAddTitleSelect(input string) {
	switch input {
	case opBack:
		break
	case opDelete:
		in.api.Data.Bookmark.Title = ""
	default:
		in.api.Data.Bookmark.Title = input
	}
	in.handleAddShow()
}

func (in *InputHandler) handleAddUrlShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup("enter a url", "", "")
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
		{Text: opDelete},
	}

	in.api.Data.State = StateAddUrlSelect
}

func (in *InputHandler) handleAddUrlSelect(input string) {
	switch input {
	case opBack:
		break
	case opDelete:
		in.api.Data.Bookmark.URL = ""
	default:
		in.api.Data.Bookmark.URL = input
	}
	in.handleAddShow()
}

func (in *InputHandler) handleAddCommentShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup("enter a comment", "", "")
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
		{Text: opDelete},
	}

	in.api.Data.State = StateAddCommentSelect
}

func (in *InputHandler) handleAddCommentSelect(input string) {
	switch input {
	case opBack:
		break
	case opDelete:
		in.api.Data.Bookmark.Comment = ""
	default:
		in.api.Data.Bookmark.Comment = input
	}
	in.handleAddShow()
}

func (in *InputHandler) handleAddTagsShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"enter some tags", "'mytag, some-tag, a tag'", "")
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
		{Text: opDelete},
	}

	in.api.Data.State = StateAddTagsSelect
}

func (in *InputHandler) handleAddTagsSelect(input string) {
	switch input {
	case opBack:
		break
	case opDelete:
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
	in.api.Data.State = StateGotoExec
	b := in.browser
	if b == "" {
		b = "xdg-open"
	}
	cmd := exec.Command(b, in.api.Data.Bookmark.URL)
	if err := cmd.Start(); err != nil {
		e := fmt.Errorf("error opening URL: %w", err)
		if b == "xdg-open" {
			e = fmt.Errorf(
				"error opening URL: xdg-utils is not installed, to use without set env variable $%s",
				robukuBrowserEnvVar)
		}
		SetMessageToError(in.api, e)
	}
}

func (in *InputHandler) handleModifyShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"select a field to edit", "", "")
	in.api.Options[rofiapi.OptionNoCustom] = "true"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	entries := []rofiapi.Entry{{Text: opBack}}
	bookmark := multiLineBookmark(in.api.Data.Bookmark)
	for _, l := range bookmark {
		entries = append(entries, rofiapi.Entry{Text: l})
	}

	in.api.Entries = entries
	in.api.Data.State = StateModifySelect
}

func (in *InputHandler) handleModifySelect(input string) {
	if len(input) < 1 {
		in.handleModifyShow()
		return
	}

	if input == opBack {
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
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"enter a new title", "", in.api.Data.Bookmark.Title)
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
		{Text: opDelete},
	}

	in.api.Data.State = StateModifyTitleSelect
}

func (in *InputHandler) handleModifyTitleSelect(input string) {
	if input == opDelete {
		input = ""
	}

	if input == opBack {
		in.handleModifyShow()
	} else if err := in.db.UpdateTitle(in.api.Data.Bookmark.ID, input); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error updating title: %w", err))
	} else {
		in.api.Data.Bookmark.Title = input
		in.handleModifyShow()
	}
}

func (in *InputHandler) handleModifyUrlShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"enter a new url", "", in.api.Data.Bookmark.URL)
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
	}

	in.api.Data.State = StateModifyUrlSelect
}

func (in *InputHandler) handleModifyUrlSelect(input string) {
	if input == "" {
		in.handleModifyUrlShow()
	} else if input == opBack {
		in.handleModifyShow()
	} else if err := in.db.UpdateURL(in.api.Data.Bookmark.ID, input); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error updating url: %w", err))
	} else {
		in.api.Data.Bookmark.URL = input
		in.handleModifyShow()
	}
}

func (in *InputHandler) handleModifyCommentShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"enter a new comment", "", in.api.Data.Bookmark.Comment)
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
		{Text: opDelete},
	}

	in.api.Data.State = StateModifyCommentSelect
}

func (in *InputHandler) handleModifyCommentSelect(input string) {
	if input == opDelete {
		input = ""
	}

	if input == opBack {
		in.handleModifyShow()
	} else if err := in.db.UpdateComment(in.api.Data.Bookmark.ID, input); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error updating comment: %w", err))
	} else {
		in.api.Data.Bookmark.Comment = input
		in.handleModifyShow()
	}
}

func (in *InputHandler) handleModifyTagsShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"add or remove tags",
		"'+ newtag1, ...' or '- oldtag1, ...'",
		strings.Join(in.api.Data.Bookmark.Tags, ", "))
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
		{Text: opDelete},
	}

	in.api.Data.State = StateModifyTagsSelect
}

func (in *InputHandler) handleModifyTagsSelect(input string) {
	switch {
	case input == opBack:
		in.handleModifyShow()
	case input == opDelete:
		if err := in.db.ClearTags(in.api.Data.Bookmark.ID); err != nil {
			SetMessageToError(in.api, fmt.Errorf("error clearing tags: %w", err))
		} else {
			in.api.Data.Bookmark.Tags = []string{}
			in.handleModifyShow()
		}
	case strings.HasPrefix(input, "+"):
		tags := getTagsFromInput(input[1:])
		if err := in.db.AddTags(in.api.Data.Bookmark.ID, tags); err != nil {
			SetMessageToError(in.api, fmt.Errorf("error adding tag: %w", err))
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
			in.handleModifyShow()
		}
	case strings.HasPrefix(input, "-"):
		tags := getTagsFromInput(input[1:])
		if err := in.db.RemoveTags(in.api.Data.Bookmark.ID, tags); err != nil {
			SetMessageToError(in.api, fmt.Errorf("error removing tag: %w", err))
		} else {
			tmp := make([]string, 0)
			for _, t := range in.api.Data.Bookmark.Tags {
				if !slices.Contains(tags, t) {
					tmp = append(tmp, t)
				}
			}
			in.api.Data.Bookmark.Tags = tmp
			in.handleModifyShow()
		}
	default:
		in.handleModifyTagsShow()
	}
}

func (in *InputHandler) handleDeleteConfirmShow() {
	in.api.Options[rofiapi.OptionMessage] = generatePangoMarkup(
		"delete? (yes/No)", "", in.api.Data.Bookmark.URL)
	in.api.Options[rofiapi.OptionNoCustom] = "false"
	in.api.Options[rofiapi.OptionUseHotKeys] = "false"

	in.api.Entries = []rofiapi.Entry{
		{Text: opBack},
	}

	in.api.Data.State = StateDeleteConfirmSelect
}

func (in *InputHandler) handleDeleteConfirmSelect(input string) {
	if input == opBack || input != "yes" {
		in.HandleBookmarksShow()
		return
	}

	if err := in.db.Remove(in.api.Data.Bookmark.ID); err != nil {
		SetMessageToError(in.api, fmt.Errorf("error deleting bookmark: %w", err))
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
func SetMessageToError(api *rofiapi.RofiApi[Data], err error) {
	log.Println("ERROR", err)
	api.Options[rofiapi.OptionMessage] = fmt.Sprintf(
		"<markup><span font_weight=\"bold\">error:</span><span> %s</span></markup>",
		rofiapi.EscapePangoMarkup(err.Error()))
	api.Options[rofiapi.OptionNoCustom] = "true"
	api.Entries = []rofiapi.Entry{{Text: opExit}}
	api.Data.State = StateErrorShow
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
		formatEntryText("> " + url),
		formatEntryText("+ " + comment),
		formatEntryText("# " + tags),
	}
}

func generatePangoMarkup(instructions, example, currentValue string) string {
	markup := "<markup>"

	if instructions != "" {
		instructions = rofiapi.EscapePangoMarkup(instructions)
		markup += fmt.Sprintf(
			"<span font_weight=\"bold\">%s</span>", instructions)
	}
	if example != "" {
		example = rofiapi.EscapePangoMarkup(example)
		if instructions != "" {
			markup += "\r"
		}
		markup += fmt.Sprintf(
			"<span font_weight=\"bold\">example:</span><span> <i>%s</i></span>",
			example)
	}
	if currentValue != "" {
		currentValue = truncateMiddle(currentValue, entryMaxLen)
		currentValue = rofiapi.EscapePangoMarkup(currentValue)
		if example != "" || instructions != "" {
			markup += "\r"
		}
		markup += fmt.Sprintf(
			"<span font_weight=\"bold\">current:</span><span> <u>%s</u></span>",
			currentValue)
	}

	markup += "</markup>"
	return markup
}

func formatEntryText(e string) string {
	e = truncateEnd(e, entryMaxLen)
	e = replaceNewlines(e)
	return e
}

func truncateMiddle(s string, l int) string {
	if len(s) > l && l >= 2 {
		half := l / 2
		return s[:half-1] + "…" + s[len(s)-half:]
	} else {
		return s
	}
}

func truncateEnd(s string, l int) string {
	if len(s) > l && l >= 0 {
		return s[0:l]
	} else {
		return s
	}
}

func replaceNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

func cleanURL(rawURL string) string {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	parsedURL.Scheme = ""
	parsedURL.Host = strings.TrimPrefix(parsedURL.Host, "www.")
	return strings.TrimPrefix(parsedURL.String(), "//")
}
