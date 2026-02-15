// Package lofigui provides a simple interface for building lightweight web UIs
// Similar to the Python version, it provides a print-like interface for generating HTML
package lofigui

import (
	"fmt"
	"html"
	"strings"
	"sync"

	"github.com/russross/blackfriday/v2"
)

// Context manages the output buffer for HTML generation
type Context struct {
	buffer        strings.Builder
	mu            sync.Mutex
	maxBufferSize int
}

// Global default context
var defaultContext = NewContext()

// NewContext creates a new Context with optional max buffer size
func NewContext() *Context {
	return &Context{
		maxBufferSize: 0, // 0 means unlimited
	}
}

// Print adds text to the buffer as HTML paragraphs
// Similar to Python's lofigui.print()
func Print(msg string, options ...PrintOption) {
	defaultContext.Print(msg, options...)
}

// Print adds text to the buffer as HTML paragraphs
func (c *Context) Print(msg string, options ...PrintOption) {
	opts := &printOptions{
		end:    "\n",
		escape: true,
	}

	for _, opt := range options {
		opt(opts)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	content := msg
	if opts.escape {
		content = html.EscapeString(msg)
	}

	if opts.end == "\n" {
		c.buffer.WriteString("<p>")
		c.buffer.WriteString(content)
		c.buffer.WriteString("</p>\n")
	} else {
		c.buffer.WriteString("&nbsp;")
		c.buffer.WriteString(content)
		c.buffer.WriteString("&nbsp;")
	}
}

// PrintOption is a functional option for Print
type PrintOption func(*printOptions)

type printOptions struct {
	end    string
	escape bool
}

// WithEnd sets the end character (use "" for inline, "\n" for paragraph)
func WithEnd(end string) PrintOption {
	return func(o *printOptions) {
		o.end = end
	}
}

// WithEscape controls HTML escaping (default true)
func WithEscape(escape bool) PrintOption {
	return func(o *printOptions) {
		o.escape = escape
	}
}

// Markdown converts markdown to HTML and adds to buffer
func Markdown(msg string) {
	defaultContext.Markdown(msg)
}

// Markdown converts markdown to HTML and adds to buffer
func (c *Context) Markdown(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	output := blackfriday.Run([]byte(msg))
	c.buffer.Write(output)
}

// HTML adds raw HTML to buffer (no escaping)
// WARNING: Only use with trusted input to avoid XSS
func HTML(msg string) {
	defaultContext.HTML(msg)
}

// HTML adds raw HTML to buffer (no escaping)
func (c *Context) HTML(msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer.WriteString(msg)
}

// Table generates an HTML table with Bulma styling
func Table(data [][]string, options ...TableOption) {
	defaultContext.Table(data, options...)
}

// Table generates an HTML table
func (c *Context) Table(data [][]string, options ...TableOption) {
	opts := &tableOptions{
		header: nil,
		escape: true,
	}

	for _, opt := range options {
		opt(opts)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer.WriteString(`<table class="table is-striped is-hoverable">`)
	c.buffer.WriteString("\n")

	// Header
	if opts.header != nil {
		c.buffer.WriteString("<thead><tr>")
		for _, h := range opts.header {
			c.buffer.WriteString("<th>")
			if opts.escape {
				c.buffer.WriteString(html.EscapeString(h))
			} else {
				c.buffer.WriteString(h)
			}
			c.buffer.WriteString("</th>")
		}
		c.buffer.WriteString("</tr></thead>\n")
	}

	// Body
	c.buffer.WriteString("<tbody>\n")
	for _, row := range data {
		c.buffer.WriteString("<tr>")
		for _, cell := range row {
			c.buffer.WriteString("<td>")
			if opts.escape {
				c.buffer.WriteString(html.EscapeString(cell))
			} else {
				c.buffer.WriteString(cell)
			}
			c.buffer.WriteString("</td>")
		}
		c.buffer.WriteString("</tr>\n")
	}
	c.buffer.WriteString("</tbody>\n")
	c.buffer.WriteString("</table>\n")
}

// TableOption is a functional option for Table
type TableOption func(*tableOptions)

type tableOptions struct {
	header []string
	escape bool
}

// WithHeader sets the table header
func WithHeader(header []string) TableOption {
	return func(o *tableOptions) {
		o.header = header
	}
}

// WithTableEscape controls HTML escaping for table cells
func WithTableEscape(escape bool) TableOption {
	return func(o *tableOptions) {
		o.escape = escape
	}
}

// Buffer returns the accumulated HTML output
func Buffer() string {
	return defaultContext.Buffer()
}

// Buffer returns the accumulated HTML output
func (c *Context) Buffer() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.buffer.String()
}

// Reset clears the buffer
func Reset() {
	defaultContext.Reset()
}

// Reset clears the buffer
func (c *Context) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.buffer.Reset()
}

// Printf is a convenience function for formatted printing
func Printf(format string, args ...interface{}) {
	Print(fmt.Sprintf(format, args...))
}

// Printf is a convenience function for formatted printing
func (c *Context) Printf(format string, args ...interface{}) {
	c.Print(fmt.Sprintf(format, args...))
}
