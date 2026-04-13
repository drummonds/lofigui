module codeberg.org/hum3/lofigui/examples/02_svg_graph

go 1.25

require (
	codeberg.org/hum3/gogal v0.1.2
	codeberg.org/hum3/lofigui v0.0.0
)

require github.com/russross/blackfriday/v2 v2.1.0 // indirect

replace codeberg.org/hum3/lofigui => ../../..
