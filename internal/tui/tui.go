package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/T4ko0522/spotify-cli/internal/config"
	"github.com/zmb3/spotify/v2"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	artistStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	stateStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type tickMsg time.Time

type Model struct {
	client      *spotify.Client
	playing     *spotify.CurrentlyPlaying
	err         error
	width       int
	quitting    bool
	showImage   bool
	albumImage  string
	lastAlbumID spotify.ID
	imgCols     int
	imgRows     int
}

func NewModel(client *spotify.Client) Model {
	return Model{
		client:    client,
		width:     80,
		showImage: IsWezTerm(),
		imgCols:   config.ImgCols,
		imgRows:   config.ImgRows,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchState, tickCmd())
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) fetchState() tea.Msg {
	playing, err := m.client.PlayerCurrentlyPlaying(context.Background())
	if err != nil {
		return errMsg{err}
	}
	return playingMsg{playing}
}

type playingMsg struct{ playing *spotify.CurrentlyPlaying }
type errMsg struct{ err error }
type imageMsg struct{ rendered string }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tickMsg:
		return m, tea.Batch(m.fetchState, tickCmd())
	case playingMsg:
		m.playing = msg.playing
		m.err = nil
		if m.showImage && msg.playing != nil && msg.playing.Item != nil {
			albumID := msg.playing.Item.Album.ID
			if albumID != m.lastAlbumID {
				m.lastAlbumID = albumID
				images := msg.playing.Item.Album.Images
				if len(images) > 0 {
					url := images[0].URL
					cols, rows := m.imgCols, m.imgRows
					return m, func() tea.Msg {
						data, err := FetchImage(url)
						if err != nil {
							return imageMsg{}
						}
						processed, err := ProcessImage(data)
						if err != nil {
							return imageMsg{rendered: RenderImageITerm2(data, cols, rows)}
						}
						return imageMsg{rendered: RenderImageITerm2(processed, cols, rows)}
					}
				}
			}
		}
	case imageMsg:
		m.albumImage = msg.rendered
		return m, tea.ClearScreen
	case errMsg:
		m.err = msg.err
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.err != nil {
		if m.showImage {
			prefix := strings.Repeat(" ", m.imgCols+2)
			return prefix + fmt.Sprintf("Error: %v", m.err) + "\n\n" + prefix + "Press q to quit.\n"
		}
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}
	if m.playing == nil || m.playing.Item == nil {
		if m.showImage {
			prefix := strings.Repeat(" ", m.imgCols+2)
			return prefix + "Nothing is currently playing." + "\n\n" + prefix + "Press q to quit.\n"
		}
		return "Nothing is currently playing.\n\nPress q to quit."
	}

	item := m.playing.Item

	state := "▶ Playing"
	if !m.playing.Playing {
		state = "⏸ Paused"
	}

	progress := int(m.playing.Progress)
	total := int(item.Duration)
	timeStr := fmt.Sprintf("%d:%02d / %d:%02d",
		progress/60000, (progress/1000)%60,
		total/60000, (total/1000)%60,
	)

	if m.showImage {
		textWidth := m.width - m.imgCols - 4
		if textWidth < 20 {
			textWidth = 20
		}
		barWidth := textWidth - len(timeStr) - 4
		if barWidth < 10 {
			barWidth = 10
		}

		t := func(s string) string { return ansi.Truncate(s, textWidth, "…") }

		var textLines []string
		switch config.ImgSize {
		case "small":
			// コンパクト: タイトル上の空行を削除
			textLines = []string{
				t(titleStyle.Render(item.Name)),
				t(artistStyle.Render(formatArtists(item.Artists))),
				"",
				t(stateStyle.Render(state)),
				"",
				t(renderProgressBar(progress, total, barWidth, timeStr)),
			}
		case "large":
			// ゆったり: スペースを多めに取る
			textLines = []string{
				"",
				"",
				t(titleStyle.Render(item.Name)),
				"",
				t(artistStyle.Render(formatArtists(item.Artists))),
				"",
				"",
				t(stateStyle.Render(state)),
				"",
				"",
				t(renderProgressBar(progress, total, barWidth, timeStr)),
			}
		default:
			// medium（デフォルト）
			textLines = []string{
				"",
				t(titleStyle.Render(item.Name)),
				t(artistStyle.Render(formatArtists(item.Artists))),
				"",
				t(stateStyle.Render(state)),
				"",
				t(renderProgressBar(progress, total, barWidth, timeStr)),
			}
		}

		var b strings.Builder

		// 画像がある場合: カーソル保存→画像描画→カーソル復元
		if m.albumImage != "" {
			b.WriteString("\033[s")      // カーソル位置を保存
			b.WriteString(m.albumImage)
			b.WriteString("\033[u")      // 保存した位置に復元（画像の左上）
		}

		// テキスト描画（常にインデント付き — 絶対カラム位置指定）
		col := m.imgCols + 3 // 1-based column: 画像幅 + gap
		limit := m.imgRows
		if limit > len(textLines) {
			limit = len(textLines)
		}
		for i := 0; i < limit; i++ {
			fmt.Fprintf(&b, "\033[%dG", col)
			line := textLines[i]
			// スペースパディングで前フレームの残像を上書き
			pad := textWidth - lipgloss.Width(line)
			if pad > 0 {
				line += strings.Repeat(" ", pad)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}

		// 画像行の残りを埋める（テキスト行が足りない場合）
		for i := len(textLines); i < m.imgRows; i++ {
			fmt.Fprintf(&b, "\033[%dG", col)
			b.WriteString(strings.Repeat(" ", textWidth))
			b.WriteString("\n")
		}

		// ヘルプテキスト（画像エリアの直下）
		b.WriteString(helpStyle.Render("Press q to quit"))
		b.WriteString("\n")

		return b.String()
	}

	// Non-image layout
	barWidth := m.width - len(timeStr) - 5
	if barWidth < 10 {
		barWidth = 10
	}

	lines := []string{
		titleStyle.Render(item.Name),
		artistStyle.Render(formatArtists(item.Artists)),
		"",
		stateStyle.Render(state),
		"",
		renderProgressBar(progress, total, barWidth, timeStr),
		"",
		helpStyle.Render("Press q to quit"),
	}

	return strings.Join(lines, "\n")
}

func formatArtists(artists []spotify.SimpleArtist) string {
	names := make([]string, len(artists))
	for i, a := range artists {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func renderProgressBar(current, total, barWidth int, timeStr string) string {
	if total == 0 {
		return "0:00 / 0:00"
	}

	filled := int(float64(current) / float64(total) * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}

	bar := strings.Repeat("#", filled) + strings.Repeat("-", barWidth-filled)
	return "[" + bar + "] " + timeStr
}

func Run(client *spotify.Client) error {
	p := tea.NewProgram(NewModel(client), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
