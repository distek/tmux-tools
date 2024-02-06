package lib

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Pane struct {
	ID     string `json:"id"`
	Index  int    `json:"index"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Active bool   `json:"active"`
}

func parsePaneLine(line string) (Pane, error) {
	var pane Pane
	var err error

	split := strings.Split(line, ",")

	if len(split) != 5 {
		return Pane{}, fmt.Errorf("lib: parsePaneLine: strings.Split: split: split length != 5: line=%s", line)
	}

	// ID
	pane.ID = split[0]

	// index
	pane.Index, err = strconv.Atoi(split[1])
	if err != nil {
		return Pane{}, fmt.Errorf("lib: parsePaneLine: strconv.Atoi: pane.Index: %s", err)
	}

	// width
	pane.Width, err = strconv.Atoi(split[2])
	if err != nil {
		return Pane{}, fmt.Errorf("lib: parsePaneLine: strconv.Atoi: pane.Width: %s", err)
	}

	// height
	pane.Height, err = strconv.Atoi(split[3])
	if err != nil {
		return Pane{}, fmt.Errorf("lib: parsePaneLine: strconv.Atoi: pane.Height: %s", err)
	}

	pane.Active = split[4] == "1"

	return pane, nil
}

func GetPanes() ([]Pane, error) {
	if UsePaneCache {
		return PaneCache, nil
	}

	UsePaneCache = true

	o, e, err := Tmux(GlobalArgs, "list-panes", map[string]string{
		"-F": "\"#{pane_id},#{pane_index},#{pane_width},#{pane_height},#{pane_active}\"",
	}, "")

	if err != nil {
		return nil, fmt.Errorf("%s: %s", err.Error(), e)
	}

	var ret []Pane

	for _, l := range strings.Split(o, "\n") {
		if l == "" || l == "\n" || l == ",,,," {
			continue
		}
		pane, err := parsePaneLine(l)
		if err != nil {
			return nil, fmt.Errorf("lib: GetPanes: parsePaneLin: pane: err=%s, line=%s", err.Error(), l)
		}

		ret = append(ret, pane)
	}

	PaneCache = ret

	return ret, nil
}

func getNeighborDirs(pane Pane) map[string]bool {
	o, _, err := Tmux(GlobalArgs, "display-message", map[string]string{
		"-p": "",
		"-t": pane.ID,
		"-F": "\"#{pane_at_left},#{pane_at_right},#{pane_at_top},#{pane_at_bottom}\"",
	}, "")
	if err != nil {
		return nil
	}

	ret := make(map[string]bool, 4)

	split := strings.Split(o, ",")

	ret["left"] = split[0] == "0"
	ret["right"] = split[1] == "0"
	ret["up"] = split[2] == "0"
	ret["down"] = split[3] == "0"

	return ret
}

func GetPaneInDir(pane Pane, dir string) (Pane, bool, error) {
	currPane, err := GetCurrentPane()
	if err != nil {
		return Pane{}, false, fmt.Errorf("lib: GetPaneInDir: getCurrentPane: ... : %s", err)
	}

	if currPane.ID != pane.ID {
		err = SelectPane(pane)
		if err != nil {
			return Pane{}, false, fmt.Errorf("lib: GetPaneInDir: SelectPane: ... : %s", err)
		}
	}

	o, e, err := Tmux(GlobalArgs, "display-message", map[string]string{
		"-p": "",
		"-t": fmt.Sprintf("{%s-of}", dir),
		"-F": "\"#{pane_id},#{pane_index},#{pane_width},#{pane_height},#{pane_active}\"",
	}, "")
	if err != nil {
		return Pane{}, false, fmt.Errorf("lib: GetPaneInDir: Tmux: command failed: err=%s, stdout=%s, stderr=%s", err, o, e)
	}

	o = strings.Split(o, "\n")[0]

	if currPane.ID != pane.ID {
		err = SelectPane(currPane)
		if err != nil {
			return Pane{}, false, fmt.Errorf("lib: GetPaneInDir: SelectPane: ... : %s", err)
		}
	}

	if o == ",,,," {
		return Pane{}, false, nil
	}

	ret, err := parsePaneLine(o)

	return ret, true, err
}

func GetCurrentPane() (Pane, error) {
	o, e, err := Tmux(GlobalArgs, "display-message", map[string]string{
		"-p": "",
		"-F": "\"#{pane_id},#{pane_index},#{pane_width},#{pane_height},#{pane_active}\"",
	}, "")

	if err != nil {
		return Pane{}, fmt.Errorf("lib: getPaneInDir: Tmux: command failed: err=%s, stdout=%s, stderr=%s", err, o, e)
	}

	return parsePaneLine(strings.Split(o, "\n")[0])
}

func SelectPane(pane Pane) error {
	o, e, err := Tmux(GlobalArgs, "select-pane", map[string]string{
		"-t": pane.ID,
	}, "")
	if err != nil {
		return fmt.Errorf("lib: selectPane: Tmux: command failed: err=%s, stdout=%s, stderr=%s", err, o, e)
	}

	return nil
}

func SwapPanes(src, dest Pane) error {
	o, e, err := Tmux(GlobalArgs, "swap-pane", map[string]string{
		"-s": src.ID,
		"-t": dest.ID,
	}, "")
	if err != nil {
		return fmt.Errorf("lib: swapPanes: Tmux: command failed: err=%s, stdout=%s, stderr=%s", err, o, e)
	}

	return nil
}

// map[dir-select][formatString, query]
var dirArgs = map[string][]string{
	"top-left":     {"\"#{pane_at_top},#{pane_at_left}\"", "11"},
	"top-right":    {"\"#{pane_at_top},#{pane_at_right}\"", "11"},
	"bottom-left":  {"\"#{pane_at_bottom},#{pane_at_left}\"", "11"},
	"bottom-right": {"\"#{pane_at_bottom},#{pane_at_right}\"", "11"},
	"left":         {"\"#{pane_at_left}\"", "1"},
	"right":        {"\"#{pane_at_right}\"", "1"},
	"up":           {"\"#{pane_at_top}\"", "1"},
	"down":         {"\"#{pane_at_bottom}\"", "1"},
}

// GetFurthestPaneInDir - valid args are:
// top-left
// top-right
// bottom-left
// bottom-right
// left
// right
// up
// down
func GetFurthestPaneInDir(dir string) (Pane, error) {
	panes, err := GetPanes()
	if err != nil {
		return Pane{}, fmt.Errorf("lib: GetFurthestPaneInDir: GetPanes: ...: %s", err)
	}

	if _, ok := dirArgs[dir]; !ok {
		return Pane{}, fmt.Errorf("lib: GetFurthestPaneInDir: GetPanes: dirArgs: direction does not exist: %s", dir)
	}

	for _, p := range panes {
		o, e, err := Tmux(GlobalArgs, "display-message", map[string]string{
			"-p": "",
			"-t": p.ID,
			"-F": dirArgs[dir][0],
		}, "")
		if err != nil {
			return Pane{}, fmt.Errorf("lib: GetFurthestPaneInDir: Tmux: command failed: err=%s, stdout=%s, stderr=%s", err, o, e)
		}

		o = strings.Split(o, "\n")[0]

		if o == dirArgs[dir][1] {
			return p, nil
		}
	}

	return Pane{}, fmt.Errorf("lib: GetFurthestPaneInDir: could not find furthest %s pane?", dir)
}

type Neighbor struct {
	Pane   Pane
	Exists bool
}

type Neighbors struct {
	Panes map[string]Neighbor
}

// Get neighboring panes ("left", "down", "up", "right")
func GetNeighbors(pane Pane) (Neighbors, error) {
	ret := Neighbors{}

	ret.Panes = make(map[string]Neighbor)

	for k, v := range getNeighborDirs(pane) {
		if !v {
			ret.Panes[k] = Neighbor{
				Pane:   Pane{},
				Exists: false,
			}

			continue
		}

		// Shadow outer var cause Go~
		k := k

		p, ok, err := GetPaneInDir(pane, k)
		if err != nil {
			return Neighbors{}, fmt.Errorf("lib: GetNeighbors: GetPaneInDir: %s : %s", k, err)
		}

		ret.Panes[k] = Neighbor{
			Pane:   p,
			Exists: ok,
		}
	}

	return ret, nil
}

// Quick func to get length of panes in current window
func GetPanesLen() int {
	panes, err := GetPanes()
	if err != nil {
		log.Println(err)
		return -1
	}

	return len(panes)
}

func KillPane(pane Pane) error {
	o, e, err := Tmux(GlobalArgs, "kill-pane", map[string]string{
		"-t": pane.ID,
	}, "")
	if err != nil {
		return fmt.Errorf("lib: KillPane: Tmux: command failed: err=%s, stdout=%s, stderr=%s", err, o, e)
	}

	return nil
}
