module github.com/drummonds/lofigui/examples/09_water_tank_htmx

go 1.21

require (
	github.com/drummonds/lofigui v0.17.0
	github.com/flosch/pongo2/v6 v6.0.0
)

require github.com/russross/blackfriday/v2 v2.1.0 // indirect

replace github.com/drummonds/lofigui => ../../..
