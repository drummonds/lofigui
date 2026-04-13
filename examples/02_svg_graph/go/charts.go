package main

import (
	"fmt"
	"math"
	"strings"
)

var chartColors = []string{"#3298dc", "#48c774", "#ffdd57", "#f14668", "#b86bff", "#ff8c42"}

// barChartSVG renders a bar chart as inline SVG.
func barChartSVG(values []float64, labels []string, title string) string {
	width, height := 500, 220
	padLeft, padRight, padTop, padBottom := 40, 20, 35, 25

	n := len(values)
	if n == 0 {
		return ""
	}

	maxVal := 0.0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	chartW := width - padLeft - padRight
	chartH := height - padTop - padBottom
	barGap := 4
	barW := (chartW - (n-1)*barGap) / n

	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, width, height)
	fmt.Fprintf(&sb, `<text x="%d" y="22" text-anchor="middle" font-size="14" font-weight="bold" fill="#333">%s</text>`,
		width/2, title)

	// Axis line
	fmt.Fprintf(&sb, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#ccc" stroke-width="1"/>`,
		padLeft, padTop+chartH, padLeft+chartW, padTop+chartH)

	for i, v := range values {
		barH := int(float64(chartH) * v / maxVal)
		if barH < 2 {
			barH = 2
		}
		x := padLeft + i*(barW+barGap)
		y := padTop + chartH - barH
		color := chartColors[i%len(chartColors)]

		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="2"/>`,
			x, y, barW, barH, color)

		// Value above bar
		fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" font-size="10" fill="#555">%.0f</text>`,
			x+barW/2, y-4, v)

		// Label below axis
		if i < len(labels) {
			fmt.Fprintf(&sb, `<text x="%d" y="%d" text-anchor="middle" font-size="10" fill="#666">%s</text>`,
				x+barW/2, height-padBottom+15, labels[i])
		}
	}

	sb.WriteString("</svg>")
	return sb.String()
}

type pieSlice struct {
	Label string
	Value float64
	Color string
}

// pieChartSVG renders a pie chart with legend as inline SVG.
func pieChartSVG(slices []pieSlice, title string) string {
	cx, cy, r := 120, 140, 90
	legendX := 260

	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg width="450" height="300" xmlns="http://www.w3.org/2000/svg">`)
	fmt.Fprintf(&sb, `<text x="120" y="25" text-anchor="middle" font-size="14" font-weight="bold" fill="#333">%s</text>`, title)

	total := 0.0
	for _, s := range slices {
		total += s.Value
	}

	startAngle := -math.Pi / 2
	for i, s := range slices {
		fraction := s.Value / total
		endAngle := startAngle + fraction*2*math.Pi

		x1 := float64(cx) + float64(r)*math.Cos(startAngle)
		y1 := float64(cy) + float64(r)*math.Sin(startAngle)
		x2 := float64(cx) + float64(r)*math.Cos(endAngle)
		y2 := float64(cy) + float64(r)*math.Sin(endAngle)

		largeArc := 0
		if fraction > 0.5 {
			largeArc = 1
		}

		fmt.Fprintf(&sb, `<path d="M%d,%d L%.1f,%.1f A%d,%d 0 %d,1 %.1f,%.1f Z" fill="%s"/>`,
			cx, cy, x1, y1, r, r, largeArc, x2, y2, s.Color)

		// Legend
		ly := 70 + i*25
		fmt.Fprintf(&sb, `<rect x="%d" y="%d" width="14" height="14" fill="%s" rx="2"/>`,
			legendX, ly, s.Color)
		fmt.Fprintf(&sb, `<text x="%d" y="%d" font-size="12" fill="#333">%s (%.0f%%)</text>`,
			legendX+20, ly+12, s.Label, fraction*100)

		startAngle = endAngle
	}

	sb.WriteString("</svg>")
	return sb.String()
}
