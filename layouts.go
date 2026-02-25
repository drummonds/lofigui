package lofigui

// Built-in Bulma layout templates for common app patterns.
// Use with NewControllerWithLayout or NewController with TemplateString.

// LayoutSingle is a minimal layout: section > container > results. No navbar, no footer.
const LayoutSingle = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{% if title %}{{ title }}{% else %}Lofigui{% endif %}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
  {{ refresh | safe }}
</head>
<body>
  <section class="section">
    <div class="container">
      {{ results | safe }}
    </div>
  </section>
</body>
</html>`

// LayoutNavbar has a fixed navbar (is-primary) with app name + status tag,
// a section with container for results, and a footer with version.
const LayoutNavbar = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{% if title %}{{ title }}{% else %}Lofigui{% endif %}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
  {{ refresh | safe }}
</head>
<body>
  <nav class="navbar is-primary" role="navigation" aria-label="main navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{ controller_name }}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        <span class="tag {% if polling == "Running" %}is-warning{% else %}is-success{% endif %}">{{ polling }}</span>
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container">
      {{ results | safe }}
    </div>
  </section>
  <footer class="footer">
    <div class="content has-text-centered">
      <p>{{ version }}</p>
    </div>
  </footer>
</body>
</html>`

// LayoutThreePanel has a navbar, a sidebar column (is-3) and main content column, plus a footer.
// Pass sidebar content via extra context key "sidebar".
const LayoutThreePanel = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{% if title %}{{ title }}{% else %}Lofigui{% endif %}</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@1.0.4/css/bulma.min.css">
  {{ refresh | safe }}
</head>
<body>
  <nav class="navbar is-primary" role="navigation" aria-label="main navigation">
    <div class="navbar-brand">
      <span class="navbar-item has-text-weight-bold">{{ controller_name }}</span>
    </div>
    <div class="navbar-end">
      <div class="navbar-item">
        <span class="tag {% if polling == "Running" %}is-warning{% else %}is-success{% endif %}">{{ polling }}</span>
      </div>
    </div>
  </nav>
  <section class="section">
    <div class="container">
      <div class="columns">
        <div class="column is-3">
          <div class="box">
            {{ sidebar | safe }}
          </div>
        </div>
        <div class="column">
          {{ results | safe }}
        </div>
      </div>
    </div>
  </section>
  <footer class="footer">
    <div class="content has-text-centered">
      <p>{{ version }}</p>
    </div>
  </footer>
</body>
</html>`

// NewControllerWithLayout creates a Controller from a built-in layout template.
//
// Example:
//
//	ctrl, err := lofigui.NewControllerWithLayout(lofigui.LayoutNavbar, "My App")
func NewControllerWithLayout(layout string, name string) (*Controller, error) {
	return NewController(ControllerConfig{
		TemplateString: layout,
		Name:           name,
	})
}
