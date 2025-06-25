package main

import (
	"testing"
	pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

func TestBufferDiffPluginGetCommands(t *testing.T) {
	plugin := &BufferDiffPlugin{}
	
	commands := plugin.GetCommands()
	
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
	
	// Test buffer-diff command
	bufferDiffCmd := findCommand(commands, "buffer-diff")
	if bufferDiffCmd == nil {
		t.Fatal("buffer-diff command not found")
	}
	
	if len(bufferDiffCmd.ArgPrompts) != 2 {
		t.Errorf("Expected 2 ArgPrompts for buffer-diff, got %d", len(bufferDiffCmd.ArgPrompts))
	}
	
	expectedPrompts := []string{"Compare buffer: ", "With buffer: "}
	for i, expected := range expectedPrompts {
		if i >= len(bufferDiffCmd.ArgPrompts) || bufferDiffCmd.ArgPrompts[i] != expected {
			t.Errorf("Expected prompt %d to be '%s', got '%s'", i, expected, bufferDiffCmd.ArgPrompts[i])
		}
	}
	
	// Test buffer-diff-current command
	bufferDiffCurrentCmd := findCommand(commands, "buffer-diff-current")
	if bufferDiffCurrentCmd == nil {
		t.Fatal("buffer-diff-current command not found")
	}
	
	if len(bufferDiffCurrentCmd.ArgPrompts) != 1 {
		t.Errorf("Expected 1 ArgPrompt for buffer-diff-current, got %d", len(bufferDiffCurrentCmd.ArgPrompts))
	}
	
	if bufferDiffCurrentCmd.ArgPrompts[0] != "Compare current buffer with: " {
		t.Errorf("Expected prompt to be 'Compare current buffer with: ', got '%s'", bufferDiffCurrentCmd.ArgPrompts[0])
	}
}

func TestBufferDiffExecuteCommandWithoutArgs(t *testing.T) {
	plugin := &BufferDiffPlugin{}
	
	// Test buffer-diff without arguments (should fail)
	err := plugin.ExecuteCommand("buffer-diff")
	if err == nil {
		t.Error("Expected error when calling buffer-diff without arguments")
	}
	
	expectedMsg := "PLUGIN_MESSAGE:buffer-diff requires 2 buffer names"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
	
	// Test buffer-diff-current without arguments (should fail)
	err = plugin.ExecuteCommand("buffer-diff-current")
	if err == nil {
		t.Error("Expected error when calling buffer-diff-current without arguments")
	}
	
	expectedMsg = "PLUGIN_MESSAGE:buffer-diff-current requires 1 buffer name"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestBufferDiffExecuteCommandWithArgs(t *testing.T) {
	plugin := &BufferDiffPlugin{}
	
	// Mock host interface is needed for actual execution
	// For now, just test that the function exists and can be called
	// Note: This will fail without proper host setup, but tests the interface
	
	err := plugin.ExecuteCommand("buffer-diff", "buffer1", "buffer2")
	// We expect this to fail with "host is nil" since we don't have a mock host
	if err == nil {
		t.Error("Expected error due to nil host")
	}
	
	if err.Error() != "ERROR: host is nil" {
		t.Errorf("Expected 'ERROR: host is nil', got '%s'", err.Error())
	}
}

func findCommand(commands []pluginsdk.CommandSpec, name string) *pluginsdk.CommandSpec {
	for _, cmd := range commands {
		if cmd.Name == name {
			return &cmd
		}
	}
	return nil
}