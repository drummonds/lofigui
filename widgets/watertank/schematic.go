// Package watertank renders an SVG schematic of a water tank with pump,
// valve, and float switches. Used by examples 07–11.
//
// Pipe routing reflects a gravity-drained tank: the inlet enters the
// upper-left side of the tank from a vertical riser fed by the pump
// (bottom-left), and the outlet leaves the tank's lower-right side at
// floor level, running along to the valve (bottom-right) and out to drain.
//
// Symbols inspired by FUXA-SVG-Widgets (MIT):
// https://github.com/frangoteam/FUXA-SVG-Widgets
package watertank

import (
	"fmt"
	"strings"
)

// State is the snapshot rendered by Render.
type State struct {
	Level     float64 // 0–100
	PumpOn    bool
	ValveOpen bool
	Running   bool

	// Hyperlink targets for SVG <a> wrapping the pump and valve symbols.
	// Empty string omits the link.
	PumpHref  string
	ValveHref string

	// Optional maintenance overlay. MaintType="pump" or "valve" draws a
	// dashed orange ring around that symbol with a "MAINT n%" label.
	MaintType     string
	MaintProgress float64
}

// Render returns the SVG string for the given tank state.
// The viewBox is 740×420; the SVG is responsive (max-width: 740px).
func Render(s State) string {
	var b strings.Builder

	pumpFill := "#dbdbdb"
	if s.PumpOn {
		pumpFill = "#48c78e"
	}
	valveFill := "#dbdbdb"
	if s.ValveOpen {
		valveFill = "#48c78e"
	}
	waterColor := "#3e8ed0"
	if s.Level > 80 {
		waterColor = "#f14668"
	} else if s.Level > 60 {
		waterColor = "#ffe08a"
	}
	pipeInColor := "#dbdbdb"
	if s.PumpOn && s.Running {
		pipeInColor = "#3e8ed0"
	}
	pipeOutColor := "#dbdbdb"
	if s.ValveOpen && s.Level > 0 {
		pipeOutColor = "#3e8ed0"
	}

	const (
		tankX = 270.0
		tankY = 40.0
		tankW = 200.0
		tankH = 300.0
	)
	waterH := tankH * s.Level / 100
	waterY := tankY + tankH - waterH
	highY := tankY + tankH*0.05 // 95% line, y=55
	lowY := tankY + tankH*0.95  // 5% line, y=325

	pumpMaint := s.MaintType == "pump"
	valveMaint := s.MaintType == "valve"

	b.WriteString(`<svg viewBox="0 0 740 420" xmlns="http://www.w3.org/2000/svg" style="max-width:740px;width:100%;height:auto">`)
	b.WriteString(`<style>text{font-family:Arial,Helvetica,sans-serif}</style>`)

	// --- Inlet pipe: L-shape (single path) — supply → pump → riser → tank top-left side ---
	// Riser sits ~2 pipe widths off the tank wall. Outline traced clockwise
	// from the top-left corner of the top horizontal segment.
	fmt.Fprintf(&b, `<path d="M228,41 L275,41 L275,55 L242,55 L242,339 L0,339 L0,325 L228,325 Z" fill="%s" stroke="#363636" stroke-width="1"/>`, pipeInColor)
	if s.PumpOn && s.Running {
		b.WriteString(`<polygon points="250,44 260,48 250,52" fill="#fff" opacity="0.6"/>`)
	}

	// --- Outlet pipe: single horizontal — tank bottom-right side → valve → drain ---
	fmt.Fprintf(&b, `<rect x="465" y="325" width="255" height="14" rx="1" fill="%s" stroke="#363636" stroke-width="1"/>`, pipeOutColor)
	b.WriteString(`<polygon points="712,328 720,332 712,336" fill="#363636"/>`)

	// --- Tank (open at top: U-shape walls drawn over a fill rect) ---
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" fill="#f5f5f5"/>`,
		tankX, tankY, tankW, tankH)
	fmt.Fprintf(&b, `<path d="M%.0f,%.0f L%.0f,%.0f L%.0f,%.0f L%.0f,%.0f" fill="none" stroke="#363636" stroke-width="3" stroke-linejoin="round"/>`,
		tankX, tankY, tankX, tankY+tankH, tankX+tankW, tankY+tankH, tankX+tankW, tankY)
	if waterH > 0.5 {
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.1f" width="%.0f" height="%.1f" fill="%s" opacity="0.7" rx="3"/>`,
			tankX+3, waterY, tankW-6, waterH, waterColor)
	}

	// Float-switch marks are drawn inside the tank wall (the inlet pipe at
	// the top-left and outlet pipe at the bottom-right would otherwise
	// collide with marks on the outside).
	fmt.Fprintf(&b, `<line x1="%.0f" y1="%.1f" x2="%.0f" y2="%.1f" stroke="#f14668" stroke-width="2" stroke-dasharray="4,2"/>`,
		tankX+4, highY, tankX+18, highY)
	fmt.Fprintf(&b, `<text x="%.0f" y="%.1f" text-anchor="start" font-size="10" fill="#f14668">95%%</text>`,
		tankX+22, highY+4)
	fmt.Fprintf(&b, `<line x1="%.0f" y1="%.1f" x2="%.0f" y2="%.1f" stroke="#b5890a" stroke-width="2" stroke-dasharray="4,2"/>`,
		tankX+4, lowY, tankX+18, lowY)
	fmt.Fprintf(&b, `<text x="%.0f" y="%.1f" text-anchor="start" font-size="10" fill="#b5890a">5%%</text>`,
		tankX+22, lowY+4)

	fmt.Fprintf(&b, `<text x="%.0f" y="195" text-anchor="middle" font-size="32" font-weight="bold" fill="#363636">%.1f%%</text>`,
		tankX+tankW/2, s.Level)
	fmt.Fprintf(&b, `<text x="%.0f" y="220" text-anchor="middle" font-size="13" fill="#4a4a4a">TANK</text>`,
		tankX+tankW/2)

	// --- Pump (bottom-left, ISA centrifugal) ---
	if s.PumpHref != "" {
		fmt.Fprintf(&b, `<a href="%s" style="cursor:pointer">`, s.PumpHref)
	}
	fmt.Fprintf(&b, `<circle cx="80" cy="332" r="35" fill="%s" stroke="#363636" stroke-width="2.5"/>`, pumpFill)
	b.WriteString(`<polygon points="67,317 67,347 97,332" fill="none" stroke="#363636" stroke-width="2"/>`)
	b.WriteString(`<text x="80" y="384" text-anchor="middle" font-size="13" font-weight="bold" fill="#363636">PUMP</text>`)
	pumpLabel := "OFF"
	if s.PumpOn {
		pumpLabel = "ON"
	}
	fmt.Fprintf(&b, `<text x="80" y="400" text-anchor="middle" font-size="11" fill="#4a4a4a">%s</text>`, pumpLabel)
	if s.PumpHref != "" {
		b.WriteString(`</a>`)
	}
	if pumpMaint {
		b.WriteString(`<circle cx="80" cy="332" r="43" fill="none" stroke="#ff9900" stroke-width="3" stroke-dasharray="8,4"/>`)
		fmt.Fprintf(&b, `<text x="80" y="416" text-anchor="middle" font-size="10" font-weight="bold" fill="#ff9900">MAINT %.0f%%</text>`, s.MaintProgress)
	}

	// --- Valve (bottom-right, ISA gate valve / bowtie) ---
	if s.ValveHref != "" {
		fmt.Fprintf(&b, `<a href="%s" style="cursor:pointer">`, s.ValveHref)
	}
	fmt.Fprintf(&b, `<polygon points="570,307 610,332 570,357" fill="%s" stroke="#363636" stroke-width="2"/>`, valveFill)
	fmt.Fprintf(&b, `<polygon points="650,307 610,332 650,357" fill="%s" stroke="#363636" stroke-width="2"/>`, valveFill)
	b.WriteString(`<text x="610" y="380" text-anchor="middle" font-size="13" font-weight="bold" fill="#363636">VALVE</text>`)
	valveLabel := "CLOSED"
	if s.ValveOpen {
		valveLabel = "OPEN"
	}
	fmt.Fprintf(&b, `<text x="610" y="396" text-anchor="middle" font-size="11" fill="#4a4a4a">%s</text>`, valveLabel)
	if s.ValveHref != "" {
		b.WriteString(`</a>`)
	}
	if valveMaint {
		b.WriteString(`<circle cx="610" cy="332" r="43" fill="none" stroke="#ff9900" stroke-width="3" stroke-dasharray="8,4"/>`)
		fmt.Fprintf(&b, `<text x="610" y="416" text-anchor="middle" font-size="10" font-weight="bold" fill="#ff9900">MAINT %.0f%%</text>`, s.MaintProgress)
	}

	if s.ValveOpen && s.Level > 0 {
		b.WriteString(`<polygon points="670,328 680,332 670,336" fill="#fff" opacity="0.6"/>`)
	}

	b.WriteString(`</svg>`)
	return b.String()
}
