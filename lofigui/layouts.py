"""Built-in Bulma layout templates for common app patterns (Jinja2)."""

LAYOUT_SINGLE = """<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ title or "Lofigui" }}</title>
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
</html>"""

LAYOUT_NAVBAR = """<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ title or "Lofigui" }}</title>
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
</html>"""

LAYOUT_THREE_PANEL = """<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ title or "Lofigui" }}</title>
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
</html>"""
