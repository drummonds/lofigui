"""
Hello World Model for WASM execution via Pyodide
This module contains the business logic that will run in the browser
"""


def model():
    """
    Simple model that demonstrates lofigui-like output.
    Since we're running in the browser, we'll return HTML directly
    rather than using a server-side buffer.
    """
    output = []

    # Simulate lofigui print functionality
    output.append("<p>Hello from Python running as WASM in your browser!</p>")
    output.append("<p>This code is executing in a Web Worker using Pyodide.</p>")

    # Add some markdown-like content
    output.append("<h2>Key Features</h2>")
    output.append("<ul>")
    output.append("<li>Python code compiled to WebAssembly</li>")
    output.append("<li>Runs entirely in the browser - no server needed</li>")
    output.append("<li>Perfect for GitHub Pages hosting</li>")
    output.append("<li>Uses Web Workers for non-blocking execution</li>")
    output.append("</ul>")

    # Add a simple table
    output.append("<h2>Example Table</h2>")
    output.append('<table class="table is-striped is-hoverable">')
    output.append('<thead><tr><th>Name</th><th>Value</th><th>Description</th></tr></thead>')
    output.append('<tbody>')
    output.append('<tr><td>Pyodide</td><td>0.25+</td><td>Python to WASM compiler</td></tr>')
    output.append('<tr><td>lofigui</td><td>0.2.3</td><td>Lofi GUI framework</td></tr>')
    output.append('<tr><td>Bulma</td><td>0.9+</td><td>CSS framework</td></tr>')
    output.append('</tbody>')
    output.append('</table>')

    return "\n".join(output)


def advanced_model():
    """
    More advanced example showing dynamic data processing in the browser
    """
    import time

    output = []
    output.append("<h2>Advanced Processing</h2>")
    output.append(f"<p>Current time (browser local): {time.strftime('%Y-%m-%d %H:%M:%S')}</p>")

    # Simple computation
    numbers = [1, 1, 2, 3, 5, 8, 13, 21, 34, 55]
    total = sum(numbers)
    average = total / len(numbers)

    output.append(f"<p>Fibonacci sequence sum: {total}</p>")
    output.append(f"<p>Average: {average:.2f}</p>")

    # Show the computation
    output.append("<h3>Data Analysis</h3>")
    output.append('<table class="table">')
    output.append('<thead><tr><th>Index</th><th>Value</th><th>Cumulative Sum</th></tr></thead>')
    output.append('<tbody>')

    cumsum = 0
    for i, num in enumerate(numbers):
        cumsum += num
        output.append(f'<tr><td>{i}</td><td>{num}</td><td>{cumsum}</td></tr>')

    output.append('</tbody>')
    output.append('</table>')

    return "\n".join(output)


if __name__ == "__main__":
    # For local testing
    print(model())
    print("\n" + "="*50 + "\n")
    print(advanced_model())
