package main

import (
	"fmt"
	"strings"
)

type Symbols struct {
	Meter        string
	Superscript  [10]string
	GraphSymbols map[string][]string
}

var defaultSymbols = Symbols{
	Meter:       "■",
	Superscript: [10]string{"⁰", "¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹"},
	GraphSymbols: map[string][]string{
		"braille_up": {
			" ", "⢀", "⢠", "⢰", "⢸",
			"⡀", "⣀", "⣠", "⣰", "⣸",
			"⡄", "⣄", "⣤", "⣴", "⣼",
			"⡆", "⣆", "⣦", "⣶", "⣾",
			"⡇", "⣇", "⣧", "⣷", "⣿",
		},
		"braille_down": {
			" ", "⠈", "⠘", "⠸", "⢸",
			"⠁", "⠉", "⠙", "⠹", "⢹",
			"⠃", "⠋", "⠛", "⠻", "⢻",
			"⠇", "⠏", "⠟", "⠿", "⢿",
			"⡇", "⡏", "⡟", "⡿", "⣿",
		},
		"block_up": {
			" ", "▗", "▗", "▐", "▐",
			"▖", "▄", "▄", "▟", "▟",
			"▖", "▄", "▄", "▟", "▟",
			"▌", "▙", "▙", "█", "█",
			"▌", "▙", "▙", "█", "█",
		},
		"block_down": {
			" ", "▝", "▝", "▐", "▐",
			"▘", "▀", "▀", "▜", "▜",
			"▘", "▀", "▀", "▜", "▜",
			"▌", "▛", "▛", "█", "█",
			"▌", "▛", "▛", "█", "█",
		},
		"tty_up": {
			" ", "░", "░", "▒", "▒",
			"░", "░", "▒", "▒", "█",
			"░", "▒", "▒", "▒", "█",
			"▒", "▒", "▒", "█", "█",
			"▒", "█", "█", "█", "█",
		},
		"tty_down": {
			" ", "░", "░", "▒", "▒",
			"░", "░", "▒", "▒", "█",
			"░", "▒", "▒", "▒", "█",
			"▒", "▒", "▒", "█", "█",
			"▒", "█", "█", "█", "█",
		},
	},
}

type BarOptions struct {
	Current    *int
	Max        *int
	Percentage *int
	Width      int
	SymbolSet  string
	ShowValue  bool
}

func renderBar(opts BarOptions) string {
	percentage := 0

	if opts.Percentage != nil {
		percentage = *opts.Percentage
	} else if opts.Current != nil && opts.Max != nil && *opts.Max != 0 {
		percentage = int((float64(*opts.Current) / float64(*opts.Max)) * 100)
	}

	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	width := opts.Width
	if width <= 0 {
		width = 10
	}

	symbolSet := opts.SymbolSet
	if symbolSet == "" {
		symbolSet = "tty_up"
	}

	symbols, ok := defaultSymbols.GraphSymbols[symbolSet]
	if !ok || len(symbols) == 0 {
		symbols = defaultSymbols.GraphSymbols["tty_up"]
	}

	var builder strings.Builder
	for cell := 0; cell < width; cell++ {
		cellStart := (cell * 100) / width
		cellEnd := ((cell + 1) * 100) / width

		level := 0
		switch {
		case percentage >= cellEnd:
			level = len(symbols) - 1
		case percentage <= cellStart:
			level = 0
		default:
			cellFill := percentage - cellStart
			cellRange := cellEnd - cellStart
			if cellRange <= 0 {
				level = 0
			} else {
				level = (cellFill * (len(symbols) - 1)) / cellRange
			}
		}

		if level < 0 {
			level = 0
		}
		if level >= len(symbols) {
			level = len(symbols) - 1
		}

		builder.WriteString(symbols[level])
	}

	if opts.ShowValue {
		builder.WriteString(fmt.Sprintf(" %d%%", percentage))
	}

	return builder.String()
}
