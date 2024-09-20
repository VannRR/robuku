// rofi, for use with https://github.com/davatorium/rofi
// also see man rofi-script
package rofi

import (
	"encoding/ascii85"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Constants for option handling and environment variable names.
const (
	optionCount int = 12
	// maxDataBytes max for env variable is 8192 in POSIX standard
	maxDataBytes int = 4096
	dataEnvVar       = "ROFI_DATA"
	stateEnvVar      = "ROFI_RETV"
)

// State represents the current state returned from Rofi.
type State byte

// String returns a string of the State variable name
func (s State) String() string {
	switch s {
	case StateInit:
		return "Stateinit"
	case StateSelected:
		return "StateSelected"
	case StateSelectedCustom:
		return "StateSelectedCustom"
	case StateCustomKeybinding1:
		return "StateCustomKeybinding1"
	case StateCustomKeybinding2:
		return "StateCustomKeybinding2"
	case StateCustomKeybinding3:
		return "StateCustomKeybinding3"
	case StateCustomKeybinding4:
		return "StateCustomKeybinding4"
	case StateCustomKeybinding5:
		return "StateCustomKeybinding5"
	case StateCustomKeybinding6:
		return "StateCustomKeybinding6"
	case StateCustomKeybinding7:
		return "StateCustomKeybinding7"
	case StateCustomKeybinding8:
		return "StateCustomKeybinding8"
	case StateCustomKeybinding9:
		return "StateCustomKeybinding9"
	case StateCustomKeybinding10:
		return "StateCustomKeybinding10"
	case StateCustomKeybinding11:
		return "StateCustomKeybinding11"
	case StateCustomKeybinding12:
		return "StateCustomKeybinding12"
	case StateCustomKeybinding13:
		return "StateCustomKeybinding13"
	case StateCustomKeybinding14:
		return "StateCustomKeybinding14"
	case StateCustomKeybinding15:
		return "StateCustomKeybinding15"
	case StateCustomKeybinding16:
		return "StateCustomKeybinding16"
	case StateCustomKeybinding17:
		return "StateCustomKeybinding17"
	case StateCustomKeybinding18:
		return "StateCustomKeybinding18"
	case StateCustomKeybinding19:
		return "StateCustomKeybinding19"
	default:
		return fmt.Sprintf("Unknown Rofi State '%d'", s)
	}
}

// Possible Rofi states.
const (
	// StateInit Initial call of script.
	StateInit State = 0

	// StateSelect Selected an entry.
	StateSelected State = 1

	// StateSelectedCustom Selected a custom entry.
	StateSelectedCustom State = 2

	// StateCustomKeybinding1, Default:  Alt+1
	StateCustomKeybinding1 State = 10

	// StateCustomKeybinding2, Default:  Alt+2
	StateCustomKeybinding2 State = 11

	// StateCustomKeybinding3, Default:  Alt+3
	StateCustomKeybinding3 State = 12

	// StateCustomKeybinding4, Default:  Alt+4
	StateCustomKeybinding4 State = 13

	// StateCustomKeybinding5, Default:  Alt+5
	StateCustomKeybinding5 State = 14

	// StateCustomKeybinding6, Default:  Alt+6
	StateCustomKeybinding6 State = 15

	// StateCustomKeybinding7, Default:  Alt+7
	StateCustomKeybinding7 State = 16

	// StateCustomKeybinding8, Default:  Alt+8
	StateCustomKeybinding8 State = 17

	// StateCustomKeybinding9, Default:  Alt+9
	StateCustomKeybinding9 State = 18

	// StateCustomKeybinding10, Default:  Alt+0
	StateCustomKeybinding10 State = 19

	// StateCustomKeybinding11, Default:  Alt+exclam
	StateCustomKeybinding11 State = 20

	// StateCustomKeybinding12, Default:  Alt+at
	StateCustomKeybinding12 State = 21

	// StateCustomKeybinding13, Default:  Alt+numbersign
	StateCustomKeybinding13 State = 22

	// StateCustomKeybinding14, Default:  Alt+dollar
	StateCustomKeybinding14 State = 23

	// StateCustomKeybinding15, Default:  Alt+percent
	StateCustomKeybinding15 State = 24

	// StateCustomKeybinding16, Default:  Alt+dead_circumflex
	StateCustomKeybinding16 State = 25

	// StateCustomKeybinding17, Default:  Alt+ampersand
	StateCustomKeybinding17 State = 26

	// StateCustomKeybinding18, Default:  Alt+asterisk
	StateCustomKeybinding18 State = 27

	// StateCustomKeybinding19, Default:  Alt+parenleft
	StateCustomKeybinding19 State = 28
)

// Option represents configuration keys for the Rofi API.
type Option string

// Available options for Rofi configuration.
const (
	// OptionPrompt updates the prompt text.
	OptionPrompt Option = "prompt"

	// OptionMessage updates the message text.
	OptionMessage Option = "message"

	// OptionMarkupRows, if set to 'true', renders markup in the row.
	OptionMarkupRows Option = "markup-rows"

	// OptionUrgent marks rows as urgent. (For syntax, see the urgent option in dmenu mode)
	OptionUrgent Option = "urgent"

	// OptionActive marks rows as active. (For syntax, see the active option in dmenu mode)
	OptionActive Option = "active"

	// OptionDelim sets the delimiter for the next rows. The default is '\n'.
	// This option should finish with this. Only call this on the first call of the script;
	// it is remembered for consecutive calls.
	OptionDelim Option = "delim"

	// OptionNoCustom, if set to 'true', only accepts listed entries and ignores custom input.
	OptionNoCustom Option = "no-custom"

	// OptionUseHotKeys, if set to 'true', enables custom keybindings for the script.
	// Warning: This breaks the normal Rofi flow.
	OptionUseHotKeys Option = "use-hot-keys"

	// OptionKeepSelection, if set, maintains the current selection position and clears the filter.
	OptionKeepSelection Option = "keep-selection"

	// OptionNewSelection, if keep-selection is set, allows you to override the selected entry (absolute position).
	OptionNewSelection Option = "new-selection"

	// OptionTheme is a small theme snippet to change the background color of a widget.
	// The theme property cannot change the interface while running; it is only usable for small changes,
	// such as the background color of widgets that get updated during display, like the row color of the listview.
	OptionTheme Option = "theme"
)

// Entry represents a selectable Rofi menu entry.
type Entry struct {
	// Text is the text the row/entry will display.
	Text string

	// Icon sets the icon for that row.
	Icon string

	// Display replaces the displayed string. The original string will still be used for filtering.
	Display string

	// Meta specifies invisible search terms used for filtering.
	Meta string

	// Info is the information that, on selection, gets placed in the ROFI_INFO environment variable.
	// This entry does not get searched for filtering.
	Info string

	// NonSelectable, if true, makes the row non-activatable.
	NonSelectable bool

	// Urgent sets the urgent flag on the entry (true/false).
	Urgent bool

	// Active sets the active flag on the entry (true/false).
	Active bool
}

// String converts an Entry struct to its Rofi string representation.
func (en Entry) String() string {
	var sb strings.Builder
	sb.WriteString(en.Text)

	if en.hasAdditionalFields() {
		sb.WriteString("\x00") // Rofi entry delimiter
	}

	if en.Icon != "" {
		sb.WriteString(fmt.Sprintf("icon\x1f%s", en.Icon))
	}
	if en.Display != "" {
		sb.WriteString(fmt.Sprintf("display\x1f%s", en.Display))
	}
	if en.Meta != "" {
		sb.WriteString(fmt.Sprintf("meta\x1f%s", en.Meta))
	}
	if en.Info != "" {
		sb.WriteString(fmt.Sprintf("info\x1f%s", en.Info))
	}
	if en.NonSelectable {
		sb.WriteString(fmt.Sprintf("nonselectable\x1f%v", en.NonSelectable))
	}
	if en.Urgent {
		sb.WriteString(fmt.Sprintf("urgent\x1f%v", en.Urgent))
	}
	if en.Active {
		sb.WriteString(fmt.Sprintf("active\x1f%v", en.Active))
	}

	return sb.String()
}

// hasAdditionalFields checks if the entry has fields beyond Text.
func (en Entry) hasAdditionalFields() bool {
	return en.Icon != "" || en.Display != "" || en.Meta != "" ||
		en.Info != "" || en.NonSelectable || en.Urgent || en.Active
}

// RofiData is an interface that requires a method to convert to a to a byte slice
// and from a byte slice.
type RofiData interface {
	Bytes() []byte
	ParseBytes([]byte) error
}

// RofiApi represents the main structure for interacting with the Rofi API.
// (RofiData type parameter should be a pointer, example: *MyStruct)
type RofiApi[T RofiData] struct {
	Options map[Option]string
	Data    T
	Entries []Entry
	state   State
}

// NewRofiApi initializes a new RofiApi instance.
// data should be a pointer that implements the RofiData interface
func NewRofiApi[T RofiData](data T) (*RofiApi[T], error) {
	stateStr := os.Getenv(stateEnvVar)
	state64, _ := strconv.ParseUint(stateStr, 10, 8)

	r := &RofiApi[T]{
		Options: make(map[Option]string),
		Entries: make([]Entry, 0),
		Data:    data,
		state:   State(state64),
	}
	if err := r.getData(); err != nil {
		return nil, err
	}
	return r, nil
}

// GetState returns rofi's current state.
func (r *RofiApi[T]) GetState() State {
	return r.state
}

// Draw updates Rofi with the current options and entries.
func (r *RofiApi[T]) Draw() error {
	if err := r.setData(); err != nil {
		return err
	}
	for k, v := range r.Options {
		fmt.Printf("\x00%s\x1f%s\n", k, v)
	}
	for _, en := range r.Entries {
		fmt.Println(en)
	}
	return nil
}

// getData retrieves data from the environment variable and decodes it.
func (r *RofiApi[T]) getData() error {
	encodedData := os.Getenv(dataEnvVar)
	if encodedData == "" {
		return nil
	}

	bytes, err := decodeASCII85(encodedData)
	if err != nil {
		return fmt.Errorf("failed to decode ASCII85 string: %w", err)
	}
	if err := r.Data.ParseBytes(bytes); err != nil {
		return fmt.Errorf("failed to convert bytes to value: %w", err)
	}

	return nil
}

// setData encodes and sets data to be passed to Rofi.
func (r *RofiApi[T]) setData() error {
	bytes := r.Data.Bytes()

	if int(float64(len(bytes))*1.25) > maxDataBytes {
		return fmt.Errorf("data byte length %d * 1.25 exceeds max allowed %d",
			len(bytes), maxDataBytes)
	}

	encodedData := encodeASCII85(bytes)

	r.Options["data"] = string(encodedData)
	return nil
}

// encodeASCII85 encodes a byte slice to ASCII85.
func encodeASCII85(data []byte) string {
	encoded := make([]byte, ascii85.MaxEncodedLen(len(data)))
	ascii85.Encode(encoded, data)
	return string(encoded)
}

// decodeASCII85 decodes an ASCII85 string to a byte slice.
func decodeASCII85(data string) ([]byte, error) {
	decoded := make([]byte, len(data))
	ndst, _, err := ascii85.Decode(decoded, []byte(data), true)
	if err != nil {
		return nil, err
	}

	trim := ndst
	for trim > ndst-4 {
		if trim == 0 {
			break
		} else if decoded[trim-1] == 0 {
			trim -= 1
		} else {
			break
		}
	}

	return decoded[:trim], nil
}
