package rofidata

import (
	"fmt"
	"strconv"
	"strings"

	"robuku/bukudb"
)

type AppState byte

const (
	St_null                  AppState = iota // 0
	St_bookmarks_show                        // 1
	St_bookmarks_select                      // 2
	St_add_show                              // 3
	St_add_select                            // 4
	St_add_title_show                        // 5
	St_add_title_select                      // 6
	St_add_url_show                          // 7
	St_add_url_select                        // 8
	St_add_comment_show                      // 9
	St_add_comment_select                    // 10
	St_add_tags_show                         // 11
	St_add_tags_select                       // 12
	St_goto_exec                             // 13
	St_modify_show                           // 14
	St_modify_select                         // 15
	St_modify_title_show                     // 16
	St_modify_title_select                   // 17
	St_modify_url_show                       // 18
	St_modify_url_select                     // 19
	St_modify_comment_show                   // 20
	St_modify_comment_select                 // 21
	St_modify_tags_show                      // 22
	St_modify_tags_select                    // 23
	St_delete_confirm_show                   // 24
	St_delete_confirm_select                 // 25
)

const (
	delim    = "\u2028"
	tagDelim = ","
)

type Data struct {
	Bookmark bukudb.Bookmark
	State    AppState
}

func (d *Data) Bytes() []byte {
	return []byte(fmt.Sprintf("%d%s%s%s%s%s%s%s%s%s%d",
		d.Bookmark.ID,
		delim,
		d.Bookmark.URL,
		delim,
		d.Bookmark.Title,
		delim,
		strings.Join(d.Bookmark.Tags, ","),
		delim,
		d.Bookmark.Comment,
		delim,
		d.State,
	))
}

func (d *Data) ParseBytes(b []byte) error {
	vals := strings.SplitN(string(b), delim, 6)
	id, err := strconv.ParseUint(vals[0], 10, 16)
	if err != nil {
		return fmt.Errorf("failed to parse ID: %w", err)
	}
	d.Bookmark.ID = uint16(id)
	d.Bookmark.URL = vals[1]
	d.Bookmark.Title = vals[2]
	d.Bookmark.Tags = strings.Split(vals[3], tagDelim)
	d.Bookmark.Comment = vals[4]
	st, err := strconv.ParseUint(vals[5], 10, 8)
	if err != nil {
		return fmt.Errorf("failed to parse State: %w", err)
	}
	d.State = AppState(st)
	return nil
}
