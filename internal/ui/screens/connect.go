// Package screens contains the main application screens for Tornado.
//
// This file implements the Connection screen - the first screen users see.
// It provides a clean, minimal interface with connection history.
package screens

import (
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/jupiterozeye/tornado/internal/assets"
	"github.com/jupiterozeye/tornado/internal/config"
	"github.com/jupiterozeye/tornado/internal/db"
	"github.com/jupiterozeye/tornado/internal/models"
	"github.com/jupiterozeye/tornado/internal/ui/styles"
)

// ConnectionState represents the current state of the connection screen
type ConnectionState int

const (
	StateWelcome ConnectionState = iota
	StateForm
	StateConnecting
)

// ConnectionItem represents a saved connection in the history list
type ConnectionItem struct {
	entry config.ConnectionEntry
}

func (i ConnectionItem) FilterValue() string { return i.entry.Name }
func (i ConnectionItem) Title() string       { return i.entry.Name }
func (i ConnectionItem) Description() string {
	return i.entry.Path
}

// ConnectModel is the model for the connection screen.
type ConnectModel struct {
	// State
	state ConnectionState

	// Form fields
	pathInput textinput.Model

	// Connection history
	showHistory    bool
	connectionList list.Model
	connections    []config.ConnectionEntry

	// UI state
	errorMsg string

	// Dimensions
	width  int
	height int

	// Styling
	styles *styles.Styles

	// Loading spinner
	spinner      spinner.Model
	spinnerFrame int
	animT        float64
	roadOffset   float64
}

type connectAnimTickMsg time.Time

func connectAnimTick() tea.Cmd {
	return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
		return connectAnimTickMsg(t)
	})
}

// NewConnectModel creates a new connection screen model.
// connections should be pre-loaded by the caller to avoid mutex contention
// on the Bubble Tea event loop goroutine.
func NewConnectModel() *ConnectModel {
	// Load connections outside the event loop is safe here (initial startup only).
	// For disconnect flow, use NewConnectModelWithConnections to avoid deadlock.
	connections := []config.ConnectionEntry{}
	if cfg := config.Get(); cfg != nil {
		connections = cfg.GetConnections()
	}
	return newConnectModelFromConnections(connections)
}

// NewConnectModelWithConnections creates a connect model with pre-loaded connections.
// Use this when switching screens mid-session to avoid holding config mutex on the
// event loop goroutine at the same time as background goroutines may hold it.
func NewConnectModelWithConnections(connections []config.ConnectionEntry) *ConnectModel {
	return newConnectModelFromConnections(connections)
}

func newConnectModelFromConnections(connections []config.ConnectionEntry) *ConnectModel {
	s := styles.Default()

	// Initialize spinner with custom parenthsis spinner
	sp := spinner.New()
	sp.Spinner = spinner.Spinner{
		Frames: []string{"⎛", "⎜", "⎝", "⎞", "⎟", "⎠"},
		FPS:    time.Second / 8,
	}
	sp.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	// Initialize form fields
	path := textinput.New()
	path.Placeholder = "/path/to/database.db"
	path.CharLimit = 256

	// Create connection history list
	var connItems []list.Item
	for _, conn := range connections {
		connItems = append(connItems, ConnectionItem{entry: conn})
	}

	connList := list.New(connItems, list.NewDefaultDelegate(), 40, 5)
	connList.Title = "Recent Connections"
	connList.SetShowStatusBar(false)
	connList.SetShowHelp(false)
	connList.SetShowPagination(false)
	connList.SetFilteringEnabled(false)
	connList.SetShowTitle(true)

	m := &ConnectModel{
		state:          StateWelcome,
		styles:         s,
		spinner:        sp,
		pathInput:      path,
		showHistory:    len(connections) > 0,
		connectionList: connList,
		connections:    connections,
	}

	// Pre-fill with most recent connection if available
	if len(connections) > 0 {
		m.pathInput.SetValue(connections[0].Path)
	}

	return m
}

// Init returns the initial command for the connection screen.
func (m *ConnectModel) Init() tea.Cmd {
	return connectAnimTick()
}

// Update handles messages for the connection screen.
func (m *ConnectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch m.state {
		case StateWelcome:
			switch msg.String() {
			case "space":
				m.state = StateForm
				m.pathInput.Focus()
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}

		case StateForm:
			return m.handleFormKeys(msg)

		case StateConnecting:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			// If there's an error showing, any key returns to form
			if m.errorMsg != "" {
				m.state = StateForm
				m.errorMsg = ""
				m.pathInput.Focus()
				return m, nil
			}
		}

	case spinner.TickMsg:
		if m.state == StateConnecting {
			m.spinnerFrame++
			return m, m.spinner.Tick
		}

	case connectAnimTickMsg:
		m.animT += 0.06
		m.roadOffset += 1.2
		return m, connectAnimTick()

	// Pass through connection messages so they bubble up to App
	case ConnectSuccessMsg:
		// Save successful connection to history
		if cfg := config.Get(); cfg != nil {
			cfg.AddConnection(m.getConfig())
		}
		return m, func() tea.Msg { return msg }

	case ConnectErrorMsg:
		m.errorMsg = msg.Err
		return m, func() tea.Msg { return msg }
	}

	// Pass messages to connection list if showing
	if m.state == StateForm && m.showHistory {
		var cmd tea.Cmd
		newListModel, cmd := m.connectionList.Update(msg)
		m.connectionList = newListModel
		return m, cmd
	}

	// Pass messages to path input
	if m.state == StateForm {
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *ConnectModel) handleFormKeys(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = StateWelcome
		m.pathInput.Blur()
		m.errorMsg = ""
		m.showHistory = false
		return m, nil

	case "ctrl+h":
		// Toggle connection history
		if len(m.connections) > 0 {
			m.showHistory = !m.showHistory
			return m, nil
		}

	case "enter":
		// Handle history list selection
		if m.showHistory {
			if item, ok := m.connectionList.SelectedItem().(ConnectionItem); ok {
				m.pathInput.SetValue(item.entry.Path)
				m.showHistory = false
				return m, nil
			}
		}

		// Connect
		return m, m.startConnection()

	default:
		// Pass to path input for editing (including paste)
		var cmd tea.Cmd
		m.pathInput, cmd = m.pathInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View renders the connection screen.
func (m *ConnectModel) View() tea.View {
	var content string
	switch m.state {
	case StateWelcome:
		content = m.viewWelcome()
	case StateForm:
		content = m.viewFormScreen()
	case StateConnecting:
		content = m.viewConnectingScreen()
	default:
		content = m.viewWelcome()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// viewFormScreen renders the form dialog in the bottom right
func (m *ConnectModel) viewFormScreen() string {
	// Render the welcome background first
	background := m.viewWelcomeBackground()

	// Render the form dialog
	dialog := m.viewForm()

	// Place dialog in bottom right corner
	return placeDialogBottomRight(background, dialog, m.width, m.height)
}

// viewConnectingScreen renders the connecting dialog in bottom right
func (m *ConnectModel) viewConnectingScreen() string {
	// Render the welcome background first
	background := m.viewWelcomeBackground()

	// Render the connecting dialog
	dialog := m.viewConnecting()

	// Place dialog in bottom right corner
	return placeDialogBottomRight(background, dialog, m.width, m.height)
}

// viewWelcomeBackground returns a solid background with logo and help (no animation)
func (m *ConnectModel) viewWelcomeBackground() string {
	logoStyle := lipgloss.NewStyle().Foreground(styles.Primary)
	logo := logoStyle.Render(assets.Logo)

	anim := m.renderTornadoAnimation()

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		MarginTop(2)
	help := helpStyle.Render("Space: Connect | Ctrl+C: Quit")

	fullLogo := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logo)
	fullHelp := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	content := lipgloss.JoinVertical(lipgloss.Left, fullLogo, anim, fullHelp)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m *ConnectModel) viewWelcome() string {
	logoStyle := lipgloss.NewStyle().Foreground(styles.Primary)
	logo := logoStyle.Render(assets.Logo)
	anim := m.renderTornadoAnimation()

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		MarginTop(2)
	help := helpStyle.Render("Space: Connect | Ctrl+C: Quit")

	fullLogo := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, logo)
	fullHelp := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, help)
	content := lipgloss.JoinVertical(lipgloss.Left, fullLogo, anim, fullHelp)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

var tornadoAnimLines = []string{
	"                                                          ",
	"        ████░░░░░░░░░░░░░░░░░░▒▒▒▒▒▒▒▒▒▒▓▓▓▓▓▓▓▓▓▓▓▓▓▓    ",
	"          ██░░▒▒▒▒░░░░░░░░░░░░▒▒▒▒▒▒▒▒▒▒▒▒▒▒▓▓▓▓▓▓██      ",
	"          ██▒▒▓▓▓▓▒▒░░░░░░▒▒▓▓▒▒▒▒▒▒▒▒▒▒▒▒▓▓▓▓▓▓░░        ",
	"            ████▓▓▒▒▒▒░░░░░░░░░░▓▓▒▒▒▒▒▒▒▒▓▓▓▓██          ",
	"                ██▓▓░░▒▒░░░░░░░░▒▒▒▒▒▒▒▒▓▓▓▓▓▓░░          ",
	"                ██▒▒▒▒▓▓░░░░░░░░░░▒▒▓▓▓▓▓▓▓▓▓▓            ",
	"                ██▓▓▓▓▓▓▒▒▒▒░░░░▓▓▓▓▒▒▓▓▓▓▓▓██            ",
	"                  ██████▓▓▒▒░░░░░░▒▒▓▓▒▒▒▒▓▓██            ",
	"                      ▓▓██▒▒▒▒░░░░░░▒▒▓▓▓▓▓▓██            ",
	"                        ██▓▓▓▓░░░░░░▓▓▒▒▒▒▓▓▓▓██░░        ",
	"                        ▓▓██▓▓▒▒▒▒▒▒░░░░▒▒▓▓▓▓▓▓██        ",
	"                          ░░██████▓▓▓▓░░░░▒▒▓▓▓▓██        ",
	"                                ░░▒▒██░░░░▒▒▓▓▓▓██░░      ",
	"                                    ██▒▒░░░░░░▒▒▓▓██      ",
	"                                    ▒▒████▒▒▓▓▓▓▓▓██      ",
	"                                      ░░▒▒██░░▒▒▓▓██░░    ",
	"                                          ██░░▒▒▒▒▓▓██    ",
	"                                          ██░░▒▒▓▓▓▓▓▓    ",
	"                                          ██░░▒▒████      ",
	"                                      ▓▓██░░▒▒▓▓██        ",
	"                                      ██░░▓▓▓▓████        ",
	"                                    ▓▓██░░▒▒▓▓██          ",
	"                                   ░██░░▒▒▓▓░░            ",
	"                                   ██▒▒▓▓██               ",
	"                                  ▓▓▓▓██                  ",
	"                                ▒▒▓▓██                    ",
	"                              ░░████                      ",
}

func (m *ConnectModel) renderTornadoAnimation() string {
	if m.width == 0 {
		return ""
	}

	n := len(tornadoAnimLines)

	// ── Style palette ─────────────────────────────────────────────
	styleCache := make(map[string]lipgloss.Style)
	stl := func(fg string) lipgloss.Style {
		if s, ok := styleCache[fg]; ok {
			return s
		}
		s := lipgloss.NewStyle().Foreground(lipgloss.Color(fg))
		styleCache[fg] = s
		return s
	}

	var (
		roadFG       = "22"      // dark green grass
		poleFG       = "94"      // brown wooden pole
		debrisFG     = "241"     // grey debris (matches tornado trail)
		treeTrunkFG  = "94"      // brown tree trunk
		treeLeafFG   = "28"      // green leaves
		houseFG      = "#BA9D8A" // brown houses
		tornadoStyle = lipgloss.NewStyle().Foreground(styles.TextMuted)
	)

	// ── Background grid ───────────────────────────────────────────
	type bgCell struct {
		r  rune
		fg string
	}
	bg := make([][]bgCell, n)
	for i := range bg {
		bg[i] = make([]bgCell, m.width)
	}
	set := func(row, col int, r rune, fg string) {
		if row >= 0 && row < n && col >= 0 && col < m.width {
			bg[row][col] = bgCell{r, fg}
		}
	}

	const (
		poleBaseSpacing = 60  // telephone poles (rare, irregular)
		treeSpacing     = 75  // trees (rare)
		houseSpacing    = 200 // houses (very rare)
		destroyFull     = 5.0
		destroyFar      = 14.0
		shakeDist       = 22.0
	)

	roadTop := n - 1 // grass
	poleTop := n - 6 // telephone poles start here and go down to road

	// Tornado ground contact X ──────────────────────────────────
	lastI := n - 1
	lastF := math.Pow(float64(lastI)/float64(maxConnectInt(1, n-1)), 1.2)
	lastAmp := 5.0*(1.0-lastF) + 2.0*lastF
	lastPhase := m.animT*3.1 + float64(lastI)*0.55
	lastSway := int(math.Sin(lastPhase) * lastAmp)

	ll := tornadoAnimLines[lastI]
	llW := lipgloss.Width(ll)
	llRunes := []rune(ll)
	firstNS, lastNS := -1, -1
	for idx, r := range llRunes {
		if r != ' ' {
			if firstNS < 0 {
				firstNS = idx
			}
			lastNS = idx
		}
	}
	if firstNS < 0 {
		firstNS, lastNS = 0, 0
	}
	artGroundCenter := (firstNS + lastNS) / 2
	tornadoGroundX := (m.width-llW)/2 + artGroundCenter + lastSway

	// Dirt road - simple brown line
	for x := 0; x < m.width; x++ {
		set(roadTop, x, '▂', roadFG)
	}

	// Uneven ground texture left of tornado (damage trail) - static, just scrolls
	groundTex := []rune{'▂', '▃', '▄', '▂', '▃', ' '}
	for x := 0; x < tornadoGroundX-2; x++ {
		hash := x*7 + int(m.roadOffset)
		if hash < 0 {
			hash = -hash
		}
		if hash%3 != 0 {
			set(roadTop, x, groundTex[hash%len(groundTex)], debrisFG)
		}
	}

	// ── Helper to draw obstacles ──────────────────────────────────
	type obstacleType int
	const (
		poleObstacle obstacleType = iota
		treeObstacle
		houseObstacle
	)

	drawObstacle := func(x int, typ obstacleType, dist float64, worldX int) {
		// Determine color based on obstacle type for rubble
		rubbleColor := debrisFG
		switch typ {
		case poleObstacle:
			rubbleColor = poleFG
		case treeObstacle:
			rubbleColor = treeTrunkFG
		case houseObstacle:
			rubbleColor = houseFG
		}

		// If left of tornado - show as destroyed rubble (static, scrolls with damage path)
		if dist < -2 {
			// More rubble for houses, less for other obstacles
			rubbleSpread := 2
			if typ == houseObstacle {
				rubbleSpread = 6 // More rubble after houses
			}
			// Scattered rubble pile - based on worldX so it's static relative to ground
			for dx := -rubbleSpread; dx <= rubbleSpread; dx++ {
				rx := x + dx
				// Static rubble based on WORLD position (worldX) so it scrolls smoothly
				shouldDraw := (worldX+dx)%7 == 0 || (worldX+dx)%11 == 3
				if typ == houseObstacle {
					shouldDraw = (worldX+dx)%3 != 0 // More dense for houses
				}
				if shouldDraw {
					set(roadTop, rx, '▒', rubbleColor)
				}
			}
			return
		}

		absDist := math.Abs(dist)

		switch typ {
		case poleObstacle:
			// Wooden telephone pole with crossbars - thicker pole
			switch {
			case absDist < destroyFull:
				// Pole being destroyed - scattered splinters, more natural
				debrisRune := []rune{'▒', '░', '│', '─', '╱', '╲'}
				for i := 0; i < 12; i++ {
					dx := (x*7+i*13)%7 - 3
					dy := (x*11 + i*17) % 6
					row := poleTop + dy
					if row <= roadTop {
						r := debrisRune[(x+i)%len(debrisRune)]
						set(row, x+dx, r, poleFG)
					}
				}
			case absDist < destroyFar:
				// Breaking pole - leans and snaps
				leanDir := 1
				if dist > 0 {
					leanDir = -1
				}
				// Bent pole leaning - thicker
				for row := poleTop + 1; row <= roadTop-1; row++ {
					leanX := x + leanDir*int((destroyFar-absDist)/destroyFar*float64(row-poleTop)/2)
					set(row, leanX, '┃', poleFG)
				}
				// Broken crossbar falling
				set(poleTop, x+leanDir, '─', poleFG)
				set(roadTop, x, '█', poleFG) // base
			case absDist < shakeDist:
				// Shaking pole - thicker
				shake := int(math.Sin(m.animT*12+float64(x)) * (shakeDist - absDist) / shakeDist * 2)
				px := x + shake
				// Crossbar at top
				set(poleTop, px-2, '─', poleFG)
				set(poleTop, px-1, '─', poleFG)
				set(poleTop, px, '┼', poleFG)
				set(poleTop, px+1, '─', poleFG)
				set(poleTop, px+2, '─', poleFG)
				// Thicker pole down to road
				for row := poleTop + 1; row <= roadTop; row++ {
					set(row, px, '┃', poleFG)
				}
			default:
				// Intact wooden telephone pole - thicker
				// Crossbar at top
				set(poleTop, x-2, '─', poleFG)
				set(poleTop, x-1, '─', poleFG)
				set(poleTop, x, '┼', poleFG)
				set(poleTop, x+1, '─', poleFG)
				set(poleTop, x+2, '─', poleFG)
				// Thicker pole extends all the way down to touch road
				for row := poleTop + 1; row <= roadTop; row++ {
					set(row, x, '┃', poleFG)
				}
			}

		case treeObstacle:
			// Draw a tree at this position
			drawTree := func(tx int, tdist float64) {
				absTDist := math.Abs(tdist)
				switch {
				case absTDist < destroyFull:
					// Tree splintering - wood colored debris
					for dx := -3; dx <= 3; dx++ {
						for dy := 0; dy < 4; dy++ {
							set(roadTop-dy, tx+dx, '░', treeTrunkFG)
						}
					}
				case absTDist < destroyFar:
					// Tree falling
					leanDir := 1
					if tdist > 0 {
						leanDir = -1
					}
					// Trunk
					set(roadTop, tx, '│', treeTrunkFG)
					set(roadTop-1, tx+leanDir, '/', treeTrunkFG)
					set(roadTop-2, tx+leanDir*2, '/', treeTrunkFG)
					// Leaves falling
					set(roadTop-2, tx+leanDir*3, '▓', treeLeafFG)
					set(roadTop-1, tx+leanDir*2, '▒', treeLeafFG)
				case absTDist < shakeDist:
					// Tree shaking
					shake := int(math.Sin(m.animT*10+float64(tx)) * (shakeDist - absTDist) / shakeDist * 1.5)
					tsx := tx + shake
					// Trunk
					set(roadTop, tsx, '│', treeTrunkFG)
					set(roadTop-1, tsx, '│', treeTrunkFG)
					// Leaves canopy
					set(roadTop-2, tsx-1, '▓', treeLeafFG)
					set(roadTop-2, tsx, '█', treeLeafFG)
					set(roadTop-2, tsx+1, '▓', treeLeafFG)
					set(roadTop-3, tsx, '▒', treeLeafFG)
				default:
					// Intact tree - brown trunk with green canopy
					// Trunk
					set(roadTop, x, '│', treeTrunkFG)
					set(roadTop-1, x, '│', treeTrunkFG)
					// Leaves canopy (bushy)
					set(roadTop-2, x-1, '▓', treeLeafFG)
					set(roadTop-2, x, '█', treeLeafFG)
					set(roadTop-2, x+1, '▓', treeLeafFG)
					set(roadTop-3, x-1, '▒', treeLeafFG)
					set(roadTop-3, x, '█', treeLeafFG)
					set(roadTop-3, x+1, '▒', treeLeafFG)
				}
			}

			// Draw the main tree
			drawTree(x, dist)

		case houseObstacle:
			// House colors
			roofFG := "#4D301B" // dark red roof
			wallFG := "#BA9D8A" // darker brown walls
			chimneyFG := "240"  // grey chimney
			doorFG := "#7F5933" // darker red door
			bushFG := "28"      // green bush

			// Draw house at this position - wider, better looking
			drawHouse := func(hx int, hdist float64) {
				absHDist := math.Abs(hdist)
				switch {
				case absHDist < destroyFull:
					// House collapsing into rubble pile - house colored debris
					for dx := -6; dx <= 6; dx++ {
						for row := roadTop - 3; row <= roadTop; row++ {
							if (hx+dx+row)%3 != 0 {
								set(row, hx+dx, '▒', wallFG)
							}
						}
					}
				case absHDist < destroyFar:
					// House damaged - roof falling, walls cracked
					// Left wall crumbling
					set(roadTop, hx-4, '█', wallFG)
					set(roadTop-1, hx-4, '█', wallFG)
					set(roadTop-2, hx-4, '▓', wallFG)
					// Right wall
					set(roadTop, hx+4, '█', wallFG)
					set(roadTop-1, hx+4, '█', wallFG)
					set(roadTop-2, hx+4, '▓', wallFG)
					// Floor
					for dx := -4; dx <= 4; dx++ {
						set(roadTop, hx+dx, '█', wallFG)
					}
					// Damaged roof with gap
					set(roadTop-2, hx-4, '▓', roofFG)
					set(roadTop-2, hx-3, '▓', roofFG)
					set(roadTop-2, hx, '░', wallFG) // hole
					set(roadTop-2, hx+3, '▓', roofFG)
					set(roadTop-2, hx+4, '▓', roofFG)
					set(roadTop-3, hx-3, '▒', roofFG)
					set(roadTop-3, hx+3, '▒', roofFG)
					set(roadTop-3, hx, '╲', wallFG) // falling debris
					// Bush on side (also damaged)
					set(roadTop, hx+6, '▓', bushFG)
				case absHDist < shakeDist:
					// House under stress - cracks appearing, slight lean, debris falling
					shake := int(math.Sin(m.animT*8+float64(hx)) * (shakeDist - absHDist) / shakeDist)
					sx := hx + shake

					// Roof shaking with slight lean toward tornado
					leanDir := 1
					if dist > 0 {
						leanDir = -1
					}
					lean := int((shakeDist - absHDist) / shakeDist * 0.5)

					// Roof peak (shaking, slightly askew)
					set(roadTop-4, sx-3+leanDir*lean, '▒', roofFG)
					set(roadTop-4, sx-2+leanDir*lean, '▓', roofFG)
					set(roadTop-4, sx-1, '▓', roofFG)
					set(roadTop-4, sx, '▓', roofFG)
					set(roadTop-4, sx+1, '▓', roofFG)
					set(roadTop-4, sx+2+leanDir*lean, '▓', roofFG)
					set(roadTop-4, sx+3+leanDir*lean, '▒', roofFG)

					// Chimney (tilted)
					set(roadTop-5, sx+2+leanDir*lean, '█', chimneyFG)
					set(roadTop-5, sx+3+leanDir*lean, '▓', chimneyFG)

					// Roof slope with stress cracks
					set(roadTop-3, sx-4, '▓', roofFG)
					set(roadTop-3, sx-3, '▓', roofFG)
					set(roadTop-3, sx-2, '░', roofFG) // crack
					set(roadTop-3, sx-1, '▓', roofFG)
					set(roadTop-3, sx, '▓', roofFG)
					set(roadTop-3, sx+1, '▓', roofFG)
					set(roadTop-3, sx+2, '▓', roofFG)
					set(roadTop-3, sx+3, '░', roofFG) // crack
					set(roadTop-3, sx+4, '▓', roofFG)

					// Solid front wall with cracks
					for row := roadTop - 2; row <= roadTop; row++ {
						for dx := -4; dx <= 4; dx++ {
							set(row, sx+dx, '█', wallFG)
						}
					}
					// Cracks in wall
					set(roadTop-1, sx-3, '░', wallFG)
					set(roadTop-2, sx+3, '░', wallFG)

					// Windows with cracks
					set(roadTop-2, sx-2, '▒', "240")
					set(roadTop-2, sx+2, '▒', "240")

					// Door (dark red)
					set(roadTop-1, sx, '▓', doorFG)
					set(roadTop, sx, '▓', doorFG)

					// Falling debris from roof
					debrisY := (int(m.animT*20) % 3)
					set(roadTop-1-debrisY, sx+5+shake, '╲', roofFG)

					// Bush shaking too
					bushShake := shake / 2
					set(roadTop, sx+6+bushShake, '▓', bushFG)
					set(roadTop, sx+7+bushShake, '▒', bushFG)
				default:
					// Intact wide house with chimney and bush
					// Roof peak (dark red)
					set(roadTop-4, hx-3, '▒', roofFG)
					set(roadTop-4, hx-2, '▓', roofFG)
					set(roadTop-4, hx-1, '▓', roofFG)
					set(roadTop-4, hx, '▓', roofFG)
					set(roadTop-4, hx+1, '▓', roofFG)
					set(roadTop-4, hx+2, '▓', roofFG)
					set(roadTop-4, hx+3, '▒', roofFG)
					// Chimney (grey)
					set(roadTop-5, hx+2, '█', chimneyFG)
					set(roadTop-5, hx+3, '█', chimneyFG)
					// Roof slope (dark red)
					set(roadTop-3, hx-4, '▓', roofFG)
					set(roadTop-3, hx-3, '▓', roofFG)
					set(roadTop-3, hx-2, '▓', roofFG)
					set(roadTop-3, hx-1, '▓', roofFG)
					set(roadTop-3, hx, '▓', roofFG)
					set(roadTop-3, hx+1, '▓', roofFG)
					set(roadTop-3, hx+2, '▓', roofFG)
					set(roadTop-3, hx+3, '▓', roofFG)
					set(roadTop-3, hx+4, '▓', roofFG)
					// Solid front wall (brown)
					for row := roadTop - 2; row <= roadTop; row++ {
						for dx := -4; dx <= 4; dx++ {
							set(row, hx+dx, '█', wallFG)
						}
					}
					// Windows (cutouts in solid wall)
					set(roadTop-2, hx-2, '░', "253")
					set(roadTop-2, hx+2, '░', "253")
					// Door (dark red)
					set(roadTop-1, hx, '▓', doorFG)
					set(roadTop, hx, '▓', doorFG)
					// Bush on side
					set(roadTop, hx+6, '▓', bushFG)
					set(roadTop, hx+7, '▒', bushFG)
				}
			}
			drawHouse(x, dist)
		}
	}

	// ── Draw all obstacles ────────────────────────────────────────
	// All scroll at same constant speed (roadOffset) - infinite generation
	scrollPos := int(m.roadOffset)

	// Wrap scrollPos to prevent integer overflow and create seamless loop
	// Use a large repeating pattern (LCM of spacings ~ 3000)
	patternRepeat := 3000
	wrappedScroll := scrollPos % patternRepeat
	if wrappedScroll < 0 {
		wrappedScroll += patternRepeat
	}

	// Generate telephone poles at irregular intervals - infinite pattern
	poleIndex := 0
	worldX := 0
	for worldX < patternRepeat {
		// Each pole has slightly different spacing (45-75)
		spacing := poleBaseSpacing - 15 + (poleIndex*23)%30
		worldX += spacing
		// Calculate screen position with wrapping
		screenX := worldX - wrappedScroll
		if screenX < -10 {
			screenX += patternRepeat
		}
		if screenX >= -10 && screenX < m.width+10 {
			drawObstacle(screenX, poleObstacle, float64(screenX-tornadoGroundX), worldX)
		}
		poleIndex++
	}

	// Generate trees at fixed world positions with offsets - infinite pattern
	for worldX := treeSpacing / 2; worldX < patternRepeat; worldX += treeSpacing {
		offset := (worldX * 17) % 25
		baseX := worldX + offset
		screenX := baseX - wrappedScroll
		if screenX < -10 {
			screenX += patternRepeat
		}
		if screenX >= -10 && screenX < m.width+10 {
			drawObstacle(screenX, treeObstacle, float64(screenX-tornadoGroundX), baseX)
			// Second tree spawns near first tree
			if (worldX*13)%4 == 0 {
				secondX := screenX + 4 + ((worldX * 7) % 5)
				// Only draw if second tree is also on screen
				if secondX >= -10 && secondX < m.width+10 {
					drawObstacle(secondX, treeObstacle, float64(secondX-tornadoGroundX), baseX+4)
				}
			}
		}
	}

	// Generate houses at fixed world positions - infinite pattern
	for worldX := houseSpacing; worldX < patternRepeat; worldX += houseSpacing {
		offset := (worldX * 23) % 50
		baseX := worldX + offset
		screenX := baseX - wrappedScroll
		if screenX < -10 {
			screenX += patternRepeat
		}
		if screenX >= -10 && screenX < m.width+10 {
			drawObstacle(screenX, houseObstacle, float64(screenX-tornadoGroundX), baseX)
		}
	}

	// ── Add tornado fog/debris at ground contact ─────────────────
	// Create random sporadic dust/debris effect at tornado base
	fogRunes := []rune{'░', '▒', '▓'}
	for i := 0; i < 15; i++ {
		// Random position around tornado base
		hash := tornadoGroundX*17 + i*31 + int(m.animT*50)
		if hash < 0 {
			hash = -hash
		}
		dx := (hash % 11) - 5 // -5 to +5
		dy := hash % 4        // 0 to 3
		x := tornadoGroundX + dx
		row := roadTop - dy
		if x >= 0 && x < m.width && row >= 0 && hash%3 != 0 {
			r := fogRunes[hash%len(fogRunes)]
			set(row, x, r, debrisFG)
		}
	}

	// ── Compose: overlay tornado art on background ────────────────
	renderCell := func(sb *strings.Builder, c bgCell) {
		if c.fg == "" {
			sb.WriteRune(' ')
		} else {
			sb.WriteString(stl(c.fg).Render(string(c.r)))
		}
	}

	var out []string
	for i, line := range tornadoAnimLines {
		funnel := math.Pow(float64(i)/float64(maxConnectInt(1, n-1)), 1.2)
		topWeight := 1.0 - funnel
		amp := 5.0*topWeight + 2.0*funnel
		phase := m.animT*3.1 + float64(i)*0.55
		sway := int(math.Sin(phase) * amp)
		lineW := lipgloss.Width(line)
		pad := (m.width-lineW)/2 + sway
		if pad < 0 {
			pad = 0
		}

		tornadoRunes := []rune(line)
		var sb strings.Builder
		col := 0

		// Pre-tornado columns
		for ; col < pad && col < m.width; col++ {
			renderCell(&sb, bg[i][col])
		}
		// Tornado art — non-space chars win; spaces show background
		for _, tr := range tornadoRunes {
			if col >= m.width {
				break
			}
			if tr != ' ' {
				sb.WriteString(tornadoStyle.Render(string(tr)))
			} else {
				renderCell(&sb, bg[i][col])
			}
			col++
		}
		// Post-tornado columns
		for ; col < m.width; col++ {
			renderCell(&sb, bg[i][col])
		}

		out = append(out, sb.String())
	}
	return strings.Join(out, "\n")
}

func padToVisualWidth(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func maxConnectInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// truncateToWidthNoEllipsis truncates a string to fit within width without adding ellipsis
func truncateToWidthNoEllipsis(s string, width int) string {
	if width < 1 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= width {
		return s
	}
	result := ""
	for _, r := range s {
		runeW := lipgloss.Width(string(r))
		if lipgloss.Width(result)+runeW > width {
			break
		}
		result += string(r)
	}
	return result
}

// spliceLineStyled replaces characters in baseLine starting at column x with overlay content,
// preserving the overlay's ANSI styling.
func spliceLineStyled(baseLine, overlay string, x, totalWidth int) string {
	if x < 0 {
		x = 0
	}
	if x >= totalWidth {
		return baseLine
	}

	overlayW := lipgloss.Width(overlay)
	if overlayW == 0 {
		return baseLine
	}

	// Calculate left padding needed to reach position x
	baseW := lipgloss.Width(baseLine)
	var leftPart string
	if baseW <= x {
		// Base line is shorter than x, pad with spaces
		leftPart = baseLine + strings.Repeat(" ", x-baseW)
	} else {
		// Truncate base line at position x (no ellipsis)
		leftPart = truncateToWidthNoEllipsis(baseLine, x)
	}

	// Calculate right part starting after the overlay
	rightStart := x + overlayW
	var rightPart string
	if rightStart < totalWidth {
		// Get the part of baseLine after the overlay position
		remaining := getFromWidth(baseLine, rightStart)
		// Pad or truncate to fill remaining width (no ellipsis)
		remainingW := lipgloss.Width(remaining)
		needed := totalWidth - rightStart
		if remainingW >= needed {
			rightPart = truncateToWidthNoEllipsis(remaining, needed)
		} else {
			rightPart = remaining + strings.Repeat(" ", needed-remainingW)
		}
	}

	return leftPart + overlay + rightPart
}

// getFromWidth returns the substring of s starting at the specified visual width
func getFromWidth(s string, startWidth int) string {
	if startWidth <= 0 {
		return s
	}
	result := ""
	currentWidth := 0
	for _, r := range s {
		runeW := lipgloss.Width(string(r))
		if currentWidth >= startWidth {
			result += string(r)
		} else if currentWidth+runeW > startWidth {
			// This rune spans the boundary, skip it
			currentWidth += runeW
		} else {
			currentWidth += runeW
		}
	}
	return result
}

func (m *ConnectModel) viewForm() string {
	bodyWidth := 50
	fieldWidth := bodyWidth - 4
	var fields []string

	// Show connection history if available and toggled
	if m.showHistory && len(m.connections) > 0 {
		m.connectionList.SetWidth(fieldWidth)
		m.connectionList.SetHeight(5)
		historySection := m.styles.Subheader.Render("Recent Connections (Ctrl+H)")
		fields = append(fields, historySection)
		fields = append(fields, m.connectionList.View())
		fields = append(fields, "")
	}

	// Path input field - just show the input without extra borders
	pathLabel := m.styles.Muted.Render("Database File:")
	pathValue := m.pathInput.View()
	fields = append(fields, pathLabel+"\n"+pathValue)

	if m.errorMsg != "" {
		fields = append(fields, m.styles.Error.Render(m.truncateError(m.errorMsg, fieldWidth)))
	}

	helpText := "enter Connect • esc Cancel"
	if len(m.connections) > 0 && !m.showHistory {
		helpText += " • ctrl+h History"
	}

	return renderDialogBox("Connect to Database", fields, helpText, bodyWidth)
}

func (m *ConnectModel) viewConnecting() string {
	body := []string{}
	if m.errorMsg != "" {
		body = append(body,
			m.styles.Error.Render("Connection Failed"),
			m.truncateError(m.errorMsg, 46),
			"",
			m.styles.Muted.Render("Press any key to return"),
		)
		return renderDialogBox("Connecting", body, "any key Back", 50)
	}

	frames := []string{"⎛", "⎜", "⎝", "⎞", "⎟", "⎠"}
	frame := frames[m.spinnerFrame%len(frames)]
	body = append(body,
		lipgloss.NewStyle().Foreground(styles.Primary).Render(frame)+"  Connecting to database...",
		"",
		m.styles.Muted.Render("Please wait"),
	)
	return renderDialogBox("Connecting", body, "esc Disabled", 50)
}

// placeDialogBottomRight places a dialog box in the bottom right corner of the screen
func placeDialogBottomRight(background, dialog string, width, height int) string {
	dialogLines := strings.Split(dialog, "\n")
	dialogWidth := 0
	for _, line := range dialogLines {
		if w := lipgloss.Width(line); w > dialogWidth {
			dialogWidth = w
		}
	}
	dialogHeight := len(dialogLines)

	// Position in bottom right with some padding
	padding := 2
	x := width - dialogWidth - padding + 2
	y := height - dialogHeight - padding

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	// Composite the dialog onto the background at position (x, y)
	baseLines := strings.Split(background, "\n")
	// Pad base to full height with empty lines (no background color)
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	for i, dialogLine := range dialogLines {
		row := y + i
		if row >= len(baseLines) {
			break
		}
		baseLines[row] = spliceLineStyled(baseLines[row], dialogLine, x, width)
	}

	return strings.Join(baseLines, "\n")
}

// spliceStringAt replaces characters in baseLine starting at column x with overlay content
func spliceStringAt(baseLine, overlay string, x, totalWidth int) string {
	if x >= totalWidth {
		return baseLine
	}

	// Pad baseLine to totalWidth if needed
	baseW := lipgloss.Width(baseLine)
	if baseW < totalWidth {
		baseLine += strings.Repeat(" ", totalWidth-baseW)
	}

	overlayW := lipgloss.Width(overlay)

	// Build result: left part of base + overlay + right part of base
	// Simple truncation without ellipsis for left part
	leftPart := ""
	w := 0
	for _, r := range baseLine {
		runeW := lipgloss.Width(string(r))
		if w+runeW > x {
			break
		}
		leftPart += string(r)
		w += runeW
	}
	rightStart := x + overlayW
	rightPart := ""
	if rightStart < totalWidth {
		// Get substring from rightStart to end
		w := 0
		start := 0
		for _, r := range baseLine {
			if w >= rightStart {
				break
			}
			start++
			w += lipgloss.Width(string(r))
		}

		result := ""
		w = 0
		for i, r := range baseLine {
			if i >= start {
				if w >= totalWidth-rightStart {
					break
				}
				result += string(r)
				w += lipgloss.Width(string(r))
			}
		}
		rightPart = result
	}

	return leftPart + overlay + rightPart
}

func (m *ConnectModel) startConnection() tea.Cmd {
	m.state = StateConnecting
	m.errorMsg = ""
	m.spinnerFrame = 0
	m.showHistory = false
	m.pathInput.Blur()

	// Get config
	config := m.getConfig()

	// Start spinner animation
	spinnerCmd := tea.Tick(time.Second/8, func(t time.Time) tea.Msg {
		return spinner.TickMsg{}
	})

	return tea.Batch(
		spinnerCmd,
		func() tea.Msg {
			// Attempt connection
			database, err := db.Open(config)
			if err != nil {
				return ConnectErrorMsg{Err: err.Error()}
			}
			return ConnectSuccessMsg{DB: database}
		},
	)
}

func (m *ConnectModel) getConfig() models.ConnectionConfig {
	return models.ConnectionConfig{
		Type: "sqlite",
		Path: m.pathInput.Value(),
	}
}

// truncateError truncates error message to fit within width without ellipsis
func (m *ConnectModel) truncateError(s string, width int) string {
	if width < 1 {
		return ""
	}
	w := lipgloss.Width(s)
	if w <= width {
		return s
	}
	// Simple truncation without ellipsis
	runes := []rune(s)
	result := ""
	for _, r := range runes {
		if lipgloss.Width(result+string(r)) > width {
			break
		}
		result += string(r)
	}
	return result
}

// Message types
type ConnectSuccessMsg struct {
	DB db.Database
}

type ConnectErrorMsg struct {
	Err string
}
