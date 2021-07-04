package util

import (
	"fmt"
	"os"
)

type Bar struct {
	percent int
	cur     int
	total   int
	rate    string
}

func NewProgress(start, total int) (bar Bar) {
	bar.cur = start + 1
	bar.total = total
	bar.percent = bar.Percent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += "█"
	}
	return bar
}

func (bar *Bar) Percent() int {
	return int(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Bar) Tick() {
	last := bar.percent
	bar.percent = bar.Percent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += "█"
	}
	fmt.Fprintf(os.Stderr, "\r[%-50s]%3d%% %6d/%d", bar.rate, bar.percent, bar.cur, bar.total)
	bar.cur++
	if bar.cur == bar.total+1 {
		fmt.Println()
	}
}
