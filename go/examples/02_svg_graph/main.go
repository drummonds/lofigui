package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/drummonds/lofigui/go/lofigui"
	"github.com/wcharczuk/go-chart/v2"
)

// Model function - contains business logic with chart rendering
func model() {
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
		Title: "Fibonacci Sequence",
		Width: 800,
		Height: 400,
		Bars: bars,
	}

	// Render to SVG
	svg := renderChartToSVG(&barChart)
	lofigui.HTML(svg)

	// Add some analysis
	lofigui.Markdown("## Analysis")
	lofigui.Printf("Sum: %d", sum(fibonacci))
	lofigui.Printf("Average: %.2f", average(fibonacci))
}

// renderChartToSVG renders a chart to SVG string
func renderChartToSVG(c chart.Chart) string {
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

// Controller manages state and routing
type Controller struct {
	templates *template.Template
}

func NewController() *Controller {
	tmpl := template.Must(template.ParseFiles("templates/hello.html"))
	return &Controller{
		templates: tmpl,
	}
}

func (ctrl *Controller) handleRoot(w http.ResponseWriter, r *http.Request) {
	// Reset buffer and run model
	lofigui.Reset()
	model()

	// Prepare template data
	data := struct {
		Results string
	}{
		Results: lofigui.Buffer(),
	}

	// Render template
	if err := ctrl.templates.ExecuteTemplate(w, "hello.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	ctrl := NewController()

	http.HandleFunc("/", ctrl.handleRoot)

	addr := ":1340"
	log.Printf("Starting server on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
