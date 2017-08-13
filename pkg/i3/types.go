package i3

// Rect represents a generic rectangular shape.
type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Output represents an i3 output (i.e. a display).
type Output struct {
	Name             string `json:"name"`
	Active           bool   `json:"active"`
	Primary          bool   `json:"primary"`
	Rect             Rect   `json:"rect"`
	CurrentWorkspace string `json:"current_workspace"`
}

// Workspace represents an i3 workspace.
type Workspace struct {
	Num     int    `json:"num"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
	Focused bool   `json:"focused"`
	Rect    Rect   `json:"rect"`
	Output  string `json:"output"`
	Urgent  bool   `json:"urgent"`
}
