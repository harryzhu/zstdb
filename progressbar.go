package sqlconf

import (
	"github.com/schollz/progressbar/v3"
)

func initProgressBar(max int64, title string) *progressbar.ProgressBar {
	if max <= 0 {
		max = -1
	}
	if title == "" {
		title = "downloading"
	}
	bar := progressbar.DefaultBytes(
		max,
		title,
	)

	return bar
}
