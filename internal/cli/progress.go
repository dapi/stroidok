package cli

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressType represents different types of progress bars
type ProgressType int

const (
	ProgressTypeBar ProgressType = iota
	ProgressTypeSpinner
	ProgressTypePercentage
	ProgressTypeBytes
)

// ProgressBarStyle defines the visual style of the progress bar
type ProgressBarStyle struct {
	Width      int
	BarChar    string
	EmptyChar  string
	LeftEnd    string
	RightEnd   string
	ShowPercent bool
	ShowCount   bool
	ShowTime    bool
	ShowSpeed   bool
}

// Default styles for progress bars
var (
	DefaultBarStyle = ProgressBarStyle{
		Width:      40,
		BarChar:    "█",
		EmptyChar:  "░",
		LeftEnd:    "[",
		RightEnd:   "]",
		ShowPercent: true,
		ShowCount:   true,
		ShowTime:    true,
		ShowSpeed:   false,
	}

	DefaultSpinnerStyle = ProgressBarStyle{
		Width:      20,
		BarChar:    "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏",
		EmptyChar:  " ",
		ShowPercent: false,
		ShowCount:   true,
		ShowTime:    true,
		ShowSpeed:   false,
	}

	DefaultBytesStyle = ProgressBarStyle{
		Width:      40,
		BarChar:    "=",
		EmptyChar:  "-",
		LeftEnd:    "[",
		RightEnd:   "]",
		ShowPercent: false,
		ShowCount:   true,
		ShowTime:    true,
		ShowSpeed:   true,
	}
)

// ProgressBar represents a customizable progress bar
type ProgressBar struct {
	mu           sync.Mutex
	total        int64
	current      int64
	style        ProgressBarStyle
	progressType ProgressType
	description  string
	startTime    time.Time
	lastUpdate   time.Time
	active       bool
	spinnerIndex int
}

// NewProgressBar creates a new progress bar
func NewProgressBar(description string, total int64) *ProgressBar {
	return NewProgressBarWithStyle(description, total, DefaultBarStyle, ProgressTypeBar)
}

// NewProgressBarWithStyle creates a new progress bar with custom style
func NewProgressBarWithStyle(description string, total int64, style ProgressBarStyle, pType ProgressType) *ProgressBar {
	return &ProgressBar{
		total:        total,
		current:      0,
		style:        style,
		progressType: pType,
		description:  description,
		startTime:    time.Now(),
		lastUpdate:   time.Now(),
		active:       false,
		spinnerIndex: 0,
	}
}

// NewSpinner creates a new spinner progress indicator
func NewSpinner(description string) *ProgressBar {
	return NewProgressBarWithStyle(description, 0, DefaultSpinnerStyle, ProgressTypeSpinner)
}

// NewBytesProgress creates a progress bar for byte operations
func NewBytesProgress(description string, totalBytes int64) *ProgressBar {
	return NewProgressBarWithStyle(description, totalBytes, DefaultBytesStyle, ProgressTypeBytes)
}

// Start starts the progress bar
func (pb *ProgressBar) Start() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.active = true
	pb.startTime = time.Now()
	pb.lastUpdate = time.Now()

	// Initial render
	pb.render()
}

// Update increments the progress by 1
func (pb *ProgressBar) Update() {
	pb.Add(1)
}

// UpdateTo sets the current progress to a specific value
func (pb *ProgressBar) UpdateTo(value int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if value < 0 {
		value = 0
	}
	if value > pb.total && pb.total > 0 {
		value = pb.total
	}

	pb.current = value
	pb.lastUpdate = time.Now()

	if pb.active {
		pb.render()
	}
}

// Add increments the progress by the specified amount
func (pb *ProgressBar) Add(delta int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current += delta
	if pb.current < 0 {
		pb.current = 0
	}
	if pb.current > pb.total && pb.total > 0 {
		pb.current = pb.total
	}

	pb.lastUpdate = time.Now()

	if pb.active {
		pb.render()
	}
}

// SetTotal updates the total value
func (pb *ProgressBar) SetTotal(total int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.total = total
	if pb.current > pb.total {
		pb.current = pb.total
	}

	if pb.active {
		pb.render()
	}
}

// IncrementTotal increments the total by the specified amount
func (pb *ProgressBar) IncrementTotal(delta int64) {
	pb.SetTotal(pb.total + delta)
}

// Description updates the progress bar description
func (pb *ProgressBar) Description(desc string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.description = desc
	if pb.active {
		pb.render()
	}
}

// Stop stops the progress bar and renders the final state
func (pb *ProgressBar) Stop() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.active = false
	pb.render()
	fmt.Println() // Move to next line after stopping
}

// Finish completes the progress bar (100%)
func (pb *ProgressBar) Finish() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.total > 0 {
		pb.current = pb.total
	}
	pb.active = false
	pb.render()
	fmt.Println() // Move to next line
}

// IsActive returns whether the progress bar is currently active
func (pb *ProgressBar) IsActive() bool {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.active
}

// GetProgress returns the current progress percentage
func (pb *ProgressBar) GetProgress() float64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.total <= 0 {
		return 0
	}
	return float64(pb.current) / float64(pb.total)
}

// render renders the progress bar
func (pb *ProgressBar) render() {
	// Move cursor to beginning of line
	fmt.Print("\r")

	var output strings.Builder

	// Add description
	if pb.description != "" {
		output.WriteString(pb.description)
		output.WriteString(" ")
	}

	switch pb.progressType {
	case ProgressTypeBar:
		output.WriteString(pb.renderBar())
	case ProgressTypeSpinner:
		output.WriteString(pb.renderSpinner())
	case ProgressTypePercentage:
		output.WriteString(pb.renderPercentage())
	case ProgressTypeBytes:
		output.WriteString(pb.renderBytes())
	default:
		output.WriteString(pb.renderBar())
	}

	fmt.Print(output.String())
}

// renderBar renders a standard progress bar
func (pb *ProgressBar) renderBar() string {
	if pb.total <= 0 {
		return fmt.Sprintf("%sProcessing...%s", pb.style.LeftEnd, pb.style.RightEnd)
	}

	percent := float64(pb.current) / float64(pb.total)
	filled := int(percent * float64(pb.style.Width))
	empty := pb.style.Width - filled

	var bar strings.Builder

	// Build progress bar
	bar.WriteString(pb.style.LeftEnd)
	for i := 0; i < filled; i++ {
		bar.WriteString(pb.style.BarChar)
	}
	for i := 0; i < empty; i++ {
		bar.WriteString(pb.style.EmptyChar)
	}
	bar.WriteString(pb.style.RightEnd)

	// Add additional information
	var info strings.Builder

	if pb.style.ShowPercent {
		info.WriteString(fmt.Sprintf(" %.1f%%", percent*100))
	}

	if pb.style.ShowCount {
		info.WriteString(fmt.Sprintf(" (%d/%d)", pb.current, pb.total))
	}

	if pb.style.ShowTime {
		elapsed := time.Since(pb.startTime)
		remaining := time.Duration(float64(elapsed) / percent * (1 - percent))
		info.WriteString(fmt.Sprintf(" ETA: %v", remaining.Round(time.Second)))
	}

	return bar.String() + info.String()
}

// renderSpinner renders a spinner
func (pb *ProgressBar) renderSpinner() string {
	if !pb.active {
		return "✓ Done"
	}

	// Get current spinner character
	spinnerChars := pb.style.BarChar
	if len(spinnerChars) == 0 {
		spinnerChars = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	}

	charIndex := (int(time.Since(pb.startTime) / 100 * time.Millisecond) % len(spinnerChars))
	if charIndex < 0 || charIndex >= len(spinnerChars) {
		charIndex = 0
	}

	var output strings.Builder
	output.WriteString(string(spinnerChars[charIndex]))

	// Add count if total is specified
	if pb.total > 0 && pb.style.ShowCount {
		output.WriteString(fmt.Sprintf(" (%d/%d)", pb.current, pb.total))
	}

	// Add elapsed time
	if pb.style.ShowTime {
		elapsed := time.Since(pb.startTime)
		output.WriteString(fmt.Sprintf(" %v", elapsed.Round(time.Second)))
	}

	return output.String()
}

// renderPercentage renders a simple percentage indicator
func (pb *ProgressBar) renderPercentage() string {
	if pb.total <= 0 {
		return "Processing..."
	}

	percent := float64(pb.current) / float64(pb.total) * 100
	return fmt.Sprintf("%.1f%% (%d/%d)", percent, pb.current, pb.total)
}

// renderBytes renders a bytes progress indicator
func (pb *ProgressBar) renderBytes() string {
	if pb.total <= 0 {
		return "Processing bytes..."
	}

	percent := float64(pb.current) / float64(pb.total)
	filled := int(percent * float64(pb.style.Width))
	empty := pb.style.Width - filled

	var bar strings.Builder

	// Build progress bar
	bar.WriteString(pb.style.LeftEnd)
	for i := 0; i < filled; i++ {
		bar.WriteString(pb.style.BarChar)
	}
	for i := 0; i < empty; i++ {
		bar.WriteString(pb.style.EmptyChar)
	}
	bar.WriteString(pb.style.RightEnd)

	// Add additional information
	var info strings.Builder

	currentFormatted := formatBytes(pb.current)
	totalFormatted := formatBytes(pb.total)

	if pb.style.ShowCount {
		info.WriteString(fmt.Sprintf(" %s/%s", currentFormatted, totalFormatted))
	}

	if pb.style.ShowTime {
		elapsed := time.Since(pb.startTime)
		info.WriteString(fmt.Sprintf(" %v", elapsed.Round(time.Second)))
	}

	if pb.style.ShowSpeed && pb.current > 0 {
		elapsed := time.Since(pb.startTime).Seconds()
		if elapsed > 0 {
			speed := float64(pb.current) / elapsed
			info.WriteString(fmt.Sprintf(" %s/s", formatBytes(int64(speed))))
		}
	}

	return bar.String() + info.String()
}

// formatBytes formats bytes into human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ProgressGroup manages multiple progress bars
type ProgressGroup struct {
	bars   []*ProgressBar
	width  int
	active bool
}

// NewProgressGroup creates a new progress group
func NewProgressGroup() *ProgressGroup {
	return &ProgressGroup{
		bars:   make([]*ProgressBar, 0),
		width:   80,
		active:  false,
	}
}

// AddBar adds a progress bar to the group
func (pg *ProgressGroup) AddBar(bar *ProgressBar) {
	pg.bars = append(pg.bars, bar)
}

// NewBar creates and adds a new progress bar to the group
func (pg *ProgressGroup) NewBar(description string, total int64) *ProgressBar {
	bar := NewProgressBar(description, total)
	pg.AddBar(bar)
	return bar
}

// NewSpinner creates and adds a new spinner to the group
func (pg *ProgressGroup) NewSpinner(description string) *ProgressBar {
	spinner := NewSpinner(description)
	pg.AddBar(spinner)
	return spinner
}

// Start starts all progress bars in the group
func (pg *ProgressGroup) Start() {
	pg.active = true
	for _, bar := range pg.bars {
		bar.Start()
	}
}

// Stop stops all progress bars in the group
func (pg *ProgressGroup) Stop() {
	pg.active = false
	for _, bar := range pg.bars {
		bar.Stop()
	}
}

// Finish finishes all progress bars in the group
func (pg *ProgressGroup) Finish() {
	pg.active = false
	for _, bar := range pg.bars {
		bar.Finish()
	}
}

// Clear clears the progress bars from screen
func (pg *ProgressGroup) Clear() {
	for i := 0; i < len(pg.bars)+1; i++ {
		fmt.Print("\r\033[K") // Clear current line
		if i < len(pg.bars) {
			fmt.Print("\033[A") // Move cursor up
		}
	}
}

// SimpleProgress creates a simple one-line progress message
func SimpleProgress(message string) {
	fmt.Printf("\r%s...", message)
}

// ClearLine clears the current line
func ClearLine() {
	fmt.Print("\r\033[K")
}