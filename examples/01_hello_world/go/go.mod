module github.com/drummonds/lofigui/examples/01_hello_world

go 1.21

require (
	github.com/drummonds/lofigui v0.0.0
	github.com/flosch/pongo2/v6 v6.0.0
)

require github.com/russross/blackfriday/v2 v2.1.0 // indirect

replace github.com/drummonds/lofigui => ../../..
