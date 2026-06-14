package cmd

import (
	"fmt"

	"github.com/T4ko0522/spotify-cli/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	selectedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type settingsModel struct {
	cursor   int
	original int
	saved    bool
	err      error
}

func newSettingsModel() settingsModel {
	cur := 0
	for i, name := range config.ImgSizeNames {
		if name == config.ImgSize {
			cur = i
			break
		}
	}
	return settingsModel{cursor: cur, original: cur}
}

func (m settingsModel) Init() tea.Cmd { return nil }

func (m settingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(config.ImgSizeNames)-1 {
				m.cursor++
			}
		case "enter":
			selected := config.ImgSizeNames[m.cursor]
			if selected != config.ImgSize {
				if err := config.SaveSettings(selected); err != nil {
					m.err = err
					return m, tea.Quit
				}
				m.saved = true
			}
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m settingsModel) View() string {
	s := "\nImage size preset:\n\n"

	for i, name := range config.ImgSizeNames {
		p := config.ImgSizePresets[name]
		label := fmt.Sprintf("%s (%dx%d)", name, p.Cols, p.Rows)

		if i == m.cursor {
			s += selectedStyle.Render("  > "+label) + "\n"
		} else {
			s += unselectedStyle.Render("    "+label) + "\n"
		}
	}

	s += "\n" + dimStyle.Render("Up/Down: select  Enter: confirm  Esc: cancel") + "\n"
	return s
}

func (a *App) newSettingsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "settings",
		Short: "Interactively change TUI settings",
		Long:  "Interactively change TUI image size preset.",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = config.Load()

			m := newSettingsModel()
			p := tea.NewProgram(m)
			result, err := p.Run()
			if err != nil {
				return err
			}

			final := result.(settingsModel)
			if final.err != nil {
				return fmt.Errorf("failed to save settings: %w", final.err)
			}
			if final.saved {
				selected := config.ImgSizeNames[final.cursor]
				preset := config.ImgSizePresets[selected]
				fmt.Printf("Settings saved: %s (%dx%d)\n", selected, preset.Cols, preset.Rows)
			} else {
				fmt.Println("No changes.")
			}
			return nil
		},
	}
}
