package tui

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/T4ko0522/spotify-cli/internal/player"
)

var (
	volTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	volHelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	volInputStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
)

type volumeModel struct {
	player    *player.Player
	volume    int
	input     string
	inputMode bool
	err       error
	done      bool
}

type volumeSetMsg struct{ volume int }
type volumeErrMsg struct{ err error }

func newVolumeModel(p *player.Player) volumeModel {
	return volumeModel{player: p}
}

func (m volumeModel) Init() tea.Cmd {
	return m.fetchVolume
}

func (m volumeModel) fetchVolume() tea.Msg {
	state, err := m.player.PlayerState(context.Background())
	if err != nil {
		return volumeErrMsg{err}
	}
	return volumeSetMsg{int(state.Device.Volume)}
}

func setVolumeCmd(p *player.Player, vol int) tea.Cmd {
	return func() tea.Msg {
		if err := p.SetVolume(context.Background(), vol); err != nil {
			return volumeErrMsg{err}
		}
		return volumeSetMsg{vol}
	}
}

func (m volumeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		// Handle input mode
		if m.inputMode {
			switch key {
			case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
				if len(m.input) < 3 {
					m.input += key
				}
				return m, nil
			case "backspace":
				if len(m.input) > 0 {
					m.input = m.input[:len(m.input)-1]
				}
				if len(m.input) == 0 {
					m.inputMode = false
				}
				return m, nil
			case "enter":
				vol, err := strconv.Atoi(m.input)
				if err != nil || vol < 0 || vol > 100 {
					m.input = ""
					m.inputMode = false
					return m, nil
				}
				m.inputMode = false
				m.input = ""
				m.volume = vol
				return m, setVolumeCmd(m.player, vol)
			case "esc":
				m.inputMode = false
				m.input = ""
				return m, nil
			}
			return m, nil
		}

		// Normal mode
		switch key {
		case "q", "esc", "ctrl+c":
			m.done = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "up", "right":
			vol := m.volume + 5
			if vol > 100 {
				vol = 100
			}
			m.volume = vol
			return m, setVolumeCmd(m.player, vol)
		case "down", "left":
			vol := m.volume - 5
			if vol < 0 {
				vol = 0
			}
			m.volume = vol
			return m, setVolumeCmd(m.player, vol)
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.inputMode = true
			m.input = key
			return m, nil
		}

	case volumeSetMsg:
		m.volume = msg.volume
		m.err = nil
	case volumeErrMsg:
		m.err = msg.err
	}
	return m, nil
}

func (m volumeModel) View() string {
	if m.done {
		return ""
	}
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	var b strings.Builder

	// Title line
	if m.inputMode {
		b.WriteString(volTitleStyle.Render(fmt.Sprintf("Volume: %d%%", m.volume)))
		b.WriteString(volInputStyle.Render(fmt.Sprintf("  → %s_", m.input)))
		b.WriteString("\n")
	} else {
		b.WriteString(volTitleStyle.Render(fmt.Sprintf("Volume: %d%%", m.volume)))
		b.WriteString("\n")
	}

	// Progress bar
	const barWidth = 25
	filled := m.volume * barWidth / 100
	bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
	fmt.Fprintf(&b, "[%s] %d/100\n", bar, m.volume)

	b.WriteString("\n")
	b.WriteString(volHelpStyle.Render("Arrows: adjust  0-9: type value  Enter: confirm  q: quit"))
	b.WriteString("\n")

	return b.String()
}

func RunVolume(p *player.Player) error {
	prog := tea.NewProgram(newVolumeModel(p))
	_, err := prog.Run()
	return err
}
