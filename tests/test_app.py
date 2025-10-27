"""Tests for the App class and controller property."""

import pytest
from lofigui.app import App
from lofigui.controller import Controller


class TestAppControllerProperty:
    """Test the controller property with safe cleanup."""

    def test_controller_can_be_set_and_retrieved(self):
        """Test that controller can be set and retrieved via property."""
        app = App()
        controller = Controller()

        assert app.controller is None

        app.controller = controller
        assert app.controller is controller

    def test_controller_can_be_cleared(self):
        """Test that controller can be set to None."""
        app = App()
        controller = Controller()

        app.controller = controller
        assert app.controller is controller

        app.controller = None
        assert app.controller is None

    def test_controller_replacement_stops_running_action(self):
        """Test that replacing controller stops any running action."""
        app = App()
        controller1 = Controller()

        # Set initial controller
        app.controller = controller1

        # Start an action
        controller1.start_action()
        assert controller1.is_action_running() is True

        # Replace with new controller - should stop old action
        controller2 = Controller()
        app.controller = controller2

        # Old controller should have action stopped
        assert controller1.is_action_running() is False
        # New controller should not have action running
        assert controller2.is_action_running() is False

    def test_controller_replacement_handles_missing_methods(self):
        """Test that controller replacement handles controllers without proper methods."""

        class MinimalController:
            """Controller without is_action_running or end_action methods."""

            pass

        app = App()
        minimal = MinimalController()

        # Should not raise error even without required methods
        app.controller = minimal
        assert app.controller is minimal

        # Should not raise error when replacing
        normal_controller = Controller()
        app.controller = normal_controller
        assert app.controller is normal_controller

    def test_controller_replacement_handles_errors_gracefully(self):
        """Test that controller replacement handles errors in cleanup gracefully."""

        class ErrorController:
            """Controller that raises errors."""

            def is_action_running(self):
                return True

            def end_action(self):
                raise RuntimeError("Cleanup failed!")

        app = App()
        error_controller = ErrorController()

        # Set controller that will error on cleanup
        app.controller = error_controller

        # Should not raise error despite end_action failing
        normal_controller = Controller()
        app.controller = normal_controller
        assert app.controller is normal_controller

    def test_controller_in_init(self):
        """Test that controller can be set during initialization."""
        controller = Controller()
        app = App(controller=controller)

        assert app.controller is controller

    def test_multiple_controller_replacements(self):
        """Test multiple controller replacements work correctly."""
        app = App()

        controllers = [Controller() for _ in range(5)]

        for i, controller in enumerate(controllers):
            controller.start_action()
            app.controller = controller

            # Only the current controller should be set
            assert app.controller is controller

            # Previous controllers should have stopped actions
            for j in range(i):
                assert controllers[j].is_action_running() is False

    def test_controller_none_to_controller(self):
        """Test setting controller when initially None."""
        app = App()  # No controller
        assert app.controller is None

        controller = Controller()
        app.controller = controller
        assert app.controller is controller

    def test_controller_to_none_stops_action(self):
        """Test setting controller to None stops running action."""
        controller = Controller()
        app = App(controller=controller)

        controller.start_action()
        assert controller.is_action_running() is True

        # Setting to None should stop action
        app.controller = None
        assert controller.is_action_running() is False

    def test_controller_setter_is_idempotent(self):
        """Test that setting the same controller again doesn't stop the action (idempotent)."""
        app = App()
        controller = Controller()

        # Set controller
        app.controller = controller
        assert app.controller is controller

        # Start an action
        controller.start_action()
        assert controller.is_action_running() is True

        # Set the same controller again - should NOT stop action (idempotent)
        app.controller = controller
        assert controller.is_action_running() is True  # Action should still be running
        assert app.controller is controller
