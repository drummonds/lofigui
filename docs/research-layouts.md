# Research: Page Layouts

Progression of page layout complexity in lofigui.

## Layout spectrum

lofigui provides built-in layout templates (`layouts.go`) that cover the common cases. Each adds one structural element to the previous.

### 1. Single page (no chrome)

The simplest layout: content in a centered container, nothing else. No header, no navigation, no footer.

```
+------------------------------------------+
|                                          |
|           {{ results | safe }}           |
|                                          |
+------------------------------------------+
```

**Template**: `LayoutSingle` — `section > container > results`.

**Use when**: quick status pages, single-purpose tools, WASM apps. The page is the output.

*lofigui examples: 01 (Hello World), 02 (SVG Graph), 03-04 (WASM).*

### 2. Single page with header bar (non-scrolling)

A fixed navbar at the top with app name and status indicator. The content scrolls beneath it. Footer shows version.

```
+==========================================+
|  App Name                     [Running]  |  <- fixed navbar
+==========================================+
|                                          |
|           {{ results | safe }}           |  <- scrolls
|                                          |
+------------------------------------------+
|              version info                |  <- footer
+------------------------------------------+
```

**Template**: `LayoutNavbar` — navbar (is-primary, fixed) + section + footer.

The navbar stays visible while content scrolls. The status tag (`Running`/`Stopped`) updates on each refresh cycle. Good for dashboards where you want the app identity and state always visible.

**Use when**: single-view dashboards, monitoring tools, long-running processes. The user needs to know what they are looking at and whether it is active.

*lofigui examples: 07 (Water Tank), 08 (Water Tank Multi-Page).*

### 3. Single page with header bar (scrolling)

Same as above but the navbar scrolls with the content — it is part of the page flow, not fixed. Use this when the header is informational but not critical to keep visible.

This is just `LayoutNavbar` without the `is-fixed-top` class. Not a separate built-in layout — modify the template string or use a custom template.

**Use when**: the header is helpful context but not essential during scrolling. Reports, logs, generated output where the user reads top-to-bottom.

### 4. Three-panel layout (header + sidebar + main)

Navbar at top, sidebar on the left (navigation, controls, filters), main content area on the right.

```
+==========================================+
|  App Name                     [Running]  |  <- fixed navbar
+==========================================+
| Sidebar    |                             |
| (is-3)     |    {{ results | safe }}     |
|            |                             |
| - Nav      |    Main content area        |
| - Controls |                             |
| - Filters  |                             |
+------------------------------------------+
|              version info                |  <- footer
+------------------------------------------+
```

**Template**: `LayoutThreePanel` — navbar + columns (sidebar is-3 + main) + footer. Pass sidebar content via the `sidebar` extra context key.

The sidebar is a Bulma `box` in a `column is-3`. It can contain navigation links, control forms, filter dropdowns, or any HTML. The main content fills the remaining `column`.

**Use when**: multi-page apps, CRUD tools with navigation, dashboards with controls alongside live data.

*lofigui examples: 09 (Water Tank HTMX with nav), 10 (Maintenance with controls).*

## Choosing a layout

| Layout | Built-in constant | Navbar | Sidebar | Typical use |
|--------|-------------------|--------|---------|-------------|
| Single page | `LayoutSingle` | No | No | Quick tools, WASM apps |
| Header (fixed) | `LayoutNavbar` | Yes (fixed) | No | Dashboards, monitoring |
| Header (scrolling) | Custom | Yes (scrolls) | No | Reports, logs |
| Three-panel | `LayoutThreePanel` | Yes (fixed) | Yes | Multi-page apps, CRUD |

## Custom layouts

For layouts beyond these patterns, use `NewController` with a `TemplateString` or `TemplatePath`. The built-in layouts are plain HTML strings — copy and modify. The only requirement is that the template includes `{{ results | safe }}` somewhere in the body, and `{{ refresh | safe }}` in the head if polling is used.

### Example: custom template from string

```go
ctrl, err := lofigui.NewController(lofigui.ControllerConfig{
    TemplateString: `<!DOCTYPE html>
<html>
<head>
  {{ refresh | safe }}
  <link rel="stylesheet" href="/assets/bulma.min.css">
</head>
<body>
  <div class="columns">
    <div class="column is-2">{{ nav | safe }}</div>
    <div class="column">{{ results | safe }}</div>
    <div class="column is-3">{{ aside | safe }}</div>
  </div>
</body>
</html>`,
    Name: "Custom Layout",
})
```

Pass extra context keys (`nav`, `aside`) via `app.HandleDisplay` or `ctrl.RenderTemplate`.
