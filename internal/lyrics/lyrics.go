package lyrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SyncedLine struct {
	TimeMs int
	Text   string
}

type Result struct {
	PlainLyrics  string
	SyncedLyrics []SyncedLine
	Synced       bool
}

type lrcResponse struct {
	PlainLyrics  string `json:"plainLyrics"`
	SyncedLyrics string `json:"syncedLyrics"`
}

func Fetch(trackName, artistName string, durationMs int) (*Result, error) {
	params := url.Values{
		"track_name":  {trackName},
		"artist_name": {artistName},
		"duration":    {strconv.Itoa(durationMs / 1000)},
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://lrclib.net/api/get?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch lyrics: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("lyrics not found")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("LRCLIB returned status %d", resp.StatusCode)
	}

	var data lrcResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	result := &Result{}

	if data.SyncedLyrics != "" {
		result.SyncedLyrics = parseLRC(data.SyncedLyrics)
		result.Synced = true
	}

	if data.PlainLyrics != "" {
		result.PlainLyrics = data.PlainLyrics
	}

	if !result.Synced && result.PlainLyrics == "" {
		return nil, fmt.Errorf("lyrics not found")
	}

	return result, nil
}

// parseLRC parses LRC format "[mm:ss.xx] text" into SyncedLine slices.
func parseLRC(lrc string) []SyncedLine {
	var lines []SyncedLine
	for _, line := range strings.Split(lrc, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "[") {
			continue
		}
		closeBracket := strings.Index(line, "]")
		if closeBracket < 0 {
			continue
		}
		tag := line[1:closeBracket]
		text := strings.TrimSpace(line[closeBracket+1:])

		ms, err := parseTimestamp(tag)
		if err != nil {
			continue
		}
		lines = append(lines, SyncedLine{TimeMs: ms, Text: text})
	}
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].TimeMs < lines[j].TimeMs
	})
	return lines
}

// parseTimestamp parses "mm:ss.xx" into milliseconds.
func parseTimestamp(ts string) (int, error) {
	parts := strings.SplitN(ts, ":", 2)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid timestamp")
	}
	min, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	secParts := strings.SplitN(parts[1], ".", 2)
	sec, err := strconv.Atoi(secParts[0])
	if err != nil {
		return 0, err
	}
	ms := 0
	if len(secParts) == 2 {
		frac := secParts[1]
		// Normalize to 3 digits
		for len(frac) < 3 {
			frac += "0"
		}
		frac = frac[:3]
		ms, err = strconv.Atoi(frac)
		if err != nil {
			return 0, err
		}
	}
	return min*60000 + sec*1000 + ms, nil
}

// CurrentLineIndex returns the index of the current lyric line based on progress.
func CurrentLineIndex(lines []SyncedLine, progressMs int) int {
	idx := -1
	for i, l := range lines {
		if l.TimeMs <= progressMs {
			idx = i
		} else {
			break
		}
	}
	return idx
}
