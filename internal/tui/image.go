package tui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
)

// IsWezTerm checks if the current terminal is WezTerm.
func IsWezTerm() bool {
	return strings.Contains(os.Getenv("TERM_PROGRAM"), "WezTerm")
}

// FetchImage downloads an image from the given URL and returns its bytes.
func FetchImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	return data, nil
}

// ProcessImage decodes, applies rounded corners, and re-encodes as PNG.
func ProcessImage(data []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	radius := w / 6

	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, bounds.Min, draw.Src)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if isOutsideRoundedCorner(x, y, w, h, radius) {
				dst.SetRGBA(x+bounds.Min.X, y+bounds.Min.Y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

func isOutsideRoundedCorner(x, y, w, h, r int) bool {
	var cx, cy int

	switch {
	case x < r && y < r:
		cx, cy = r, r
	case x >= w-r && y < r:
		cx, cy = w-r-1, r
	case x < r && y >= h-r:
		cx, cy = r, h-r-1
	case x >= w-r && y >= h-r:
		cx, cy = w-r-1, h-r-1
	default:
		return false
	}

	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy > r*r
}

// RenderImageITerm2 returns the iTerm2/WezTerm inline image escape sequence.
func RenderImageITerm2(data []byte, widthCells, heightCells int) string {
	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("\033]1337;File=inline=1;width=%d;height=%d;preserveAspectRatio=1:%s\a", widthCells, heightCells, b64)
}
