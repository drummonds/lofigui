package main

import (
	"log"
	"net/http"

	"github.com/drummonds/lofigui"
	"github.com/wcharczuk/go-chart/v2"
)

// Model function - contains business logic with chart rendering
// This is a static/synchronous example - generates immediately without polling
// The model receives the App, which manages the singleton active model state
func model(app *lofigui.App) {
	lofigui.Print("Hello to SVG graphs in Go!")

	lofigui.Markdown("## Fibonacci Bar Chart")
	lofigui.Print("Generated using go-chart library")

	// Create bar chart data
	fibonacci := []float64{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55}

	// Create bar chart
	bars := make([]chart.Value, len(fibonacci))
	for i, v := range fibonacci {
		bars[i] = chart.Value{
			Value: v,
			Label: "",
		}
	}

	barChart := chart.BarChart{
		Title:  "Fibonacci Sequence",
		Width:  800,
		Height: 400,
		Bars:   bars,
	}

	// Render to SVG
	svg := renderChartToSVG(barChart)
	lofigui.HTML(svg)

	// Add some analysis
	lofigui.Markdown("## Analysis")
	lofigui.Printf("Sum: %d", sum(fibonacci))
	lofigui.Printf("Average: %.2f", average(fibonacci))

	// End action immediately since this is synchronous (app-level state)
	app.EndAction()
}

// renderChartToSVG renders a chart to SVG string
func renderChartToSVG(c chart.BarChart) string {
	collector := &svgCollector{}
	if err := c.Render(chart.SVG, collector); err != nil {
		return "<p>Error rendering chart</p>"
	}
	return string(collector.data)
}

// svgCollector implements io.Writer to collect SVG output
type svgCollector struct {
	data []byte
}

func (s *svgCollector) Write(p []byte) (n int, err error) {
	s.data = append(s.data, p...)
	return len(p), nil
}

// Helper functions
func sum(values []float64) float64 {
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return sum(values) / float64(len(values))
}

func main() {
	// Create an App which provides safe controller management
	// The App uses composition to integrate the controller
	app := lofigui.NewApp()
	app.Version = "SVG Graph Demo v1.0"

	// Create controller using lofigui.Controller with composition pattern
	// This demonstrates how model-specific controllers are integrated via composition
	ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
		Name:         "SVG Graph Controller",    // Name displayed in app
		TemplatePath: "../templates/hello.html", // Shared template
	})
	if err != nil {
		log.Fatalf("Failed to create controller: %v", err)
	}

	// Set the controller in the app (composition pattern)
	app.SetController(ctrl)

	// Root endpoint - generates the graph immediately
	// This is a static/synchronous example - no polling needed
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		app.HandleRoot(w, r, model, true)
	})

	// Display endpoint - shows the generated graph
	http.HandleFunc("/display", func(w http.ResponseWriter, r *http.Request) {
		app.HandleDisplay(w, r)
	})

	// Favicon endpoint
	http.HandleFunc("/favicon.ico", lofigui.ServeFavicon)

	addr := ":1340"
	log.Printf("Starting server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
