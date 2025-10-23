"""Tests to ensure all examples can be imported and run correctly."""

import pytest
import sys
import os
from pathlib import Path


class TestPythonExamples:
    """Test Python examples can be imported and instantiated."""

    def test_example_01_hello_world_imports(self):
        """Test that example 01 (hello world) can be imported."""
        example_path = Path(__file__).parent.parent / "examples" / "01_hello_world" / "python"
        sys.path.insert(0, str(example_path))
        
        try:
            import hello
            
            # Verify the app exists
            assert hasattr(hello, 'app')
            assert hello.app is not None
            
            # Verify controller exists
            assert hasattr(hello, 'controller')
            assert hello.controller is not None
            
            # Verify model function exists
            assert hasattr(hello, 'model')
            assert callable(hello.model)
            
        finally:
            sys.path.remove(str(example_path))
            if 'hello' in sys.modules:
                del sys.modules['hello']

    def test_example_01_model_runs(self):
        """Test that example 01 model function runs without error."""
        example_path = Path(__file__).parent.parent / "examples" / "01_hello_world" / "python"
        sys.path.insert(0, str(example_path))
        
        try:
            import hello
            import lofigui
            
            # Reset and run model
            lofigui.reset()
            hello.model()
            
            # Check that output was generated
            output = lofigui.buffer()
            assert len(output) > 0
            assert "Hello world" in output
            
        finally:
            sys.path.remove(str(example_path))
            if 'hello' in sys.modules:
                del sys.modules['hello']

    def test_example_02_svg_graph_imports(self):
        """Test that example 02 (svg graph) can be imported."""
        example_path = Path(__file__).parent.parent / "examples" / "02_svg_graph" / "python"
        sys.path.insert(0, str(example_path))
        
        try:
            import graph
            
            # Verify the app exists
            assert hasattr(graph, 'app')
            assert graph.app is not None
            
            # Verify controller exists
            assert hasattr(graph, 'controller')
            assert graph.controller is not None
            
            # Verify model function exists
            assert hasattr(graph, 'model')
            assert callable(graph.model)
            
        finally:
            sys.path.remove(str(example_path))
            if 'graph' in sys.modules:
                del sys.modules['graph']

    def test_example_02_model_runs(self):
        """Test that example 02 model function runs without error."""
        pytest.importorskip("pygal")
        
        example_path = Path(__file__).parent.parent / "examples" / "02_svg_graph" / "python"
        sys.path.insert(0, str(example_path))
        
        try:
            import graph
            import lofigui
            
            # Reset and run model
            lofigui.reset()
            graph.model()
            
            # Check that output was generated
            output = lofigui.buffer()
            assert len(output) > 0
            assert "graph" in output.lower()
            # Should contain SVG from pygal
            assert "svg" in output.lower()
            
        finally:
            sys.path.remove(str(example_path))
            if 'graph' in sys.modules:
                del sys.modules['graph']


class TestPythonExamplesWithFastAPI:
    """Test Python examples with FastAPI test client."""

    @pytest.mark.asyncio
    async def test_example_01_fastapi_endpoint(self):
        """Test example 01 FastAPI endpoint returns valid response."""
        pytest.importorskip("fastapi")

        example_path = Path(__file__).parent.parent / "examples" / "01_hello_world" / "python"
        original_dir = os.getcwd()
        sys.path.insert(0, str(example_path))

        try:
            # Change to example directory so template paths work
            os.chdir(example_path)

            from fastapi.testclient import TestClient
            import hello
            import lofigui

            client = TestClient(hello.app)

            # Disable startup protection for testing
            hello.app.startup = False

            # Test display endpoint without action (should show initial state)
            # First manually trigger model to populate buffer
            lofigui.reset()
            lofigui.print("Test message")
            response = client.get("/display")
            assert response.status_code == 200
            assert "text/html" in response.headers["content-type"]
            assert "Test message" in response.text

            # Test favicon endpoint
            response = client.get("/favicon.ico")
            assert response.status_code == 200
            assert response.headers["content-type"] == "image/x-icon"
            assert len(response.content) > 0

        finally:
            os.chdir(original_dir)
            sys.path.remove(str(example_path))
            if 'hello' in sys.modules:
                del sys.modules['hello']

    @pytest.mark.asyncio
    async def test_example_02_fastapi_endpoint(self):
        """Test example 02 FastAPI endpoint returns valid response."""
        pytest.importorskip("fastapi")
        pytest.importorskip("pygal")
        
        example_path = Path(__file__).parent.parent / "examples" / "02_svg_graph" / "python"
        original_dir = os.getcwd()
        sys.path.insert(0, str(example_path))
        
        try:
            # Change to example directory so template paths work
            os.chdir(example_path)
            
            from fastapi.testclient import TestClient
            import graph
            
            client = TestClient(graph.app)
            
            # Test root endpoint
            response = client.get("/")
            assert response.status_code == 200
            assert "text/html" in response.headers["content-type"]
            assert "graph" in response.text.lower()
            
            # Test favicon endpoint
            response = client.get("/favicon.ico")
            assert response.status_code == 200
            assert response.headers["content-type"] == "image/x-icon"
            
        finally:
            os.chdir(original_dir)
            sys.path.remove(str(example_path))
            if 'graph' in sys.modules:
                del sys.modules['graph']
