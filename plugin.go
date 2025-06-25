package main

import (
	"context"
	"fmt"
	"net/rpc"
	"strings"

	"github.com/hashicorp/go-plugin"
	pluginsdk "github.com/TakahashiShuuhei/gmacs-plugin-sdk"
)

type BufferDiffPlugin struct {
	host pluginsdk.HostInterface
}

func (p *BufferDiffPlugin) Name() string {
	return "buffer-diff-plugin"
}

func (p *BufferDiffPlugin) Version() string {
	return "1.0.0"
}

func (p *BufferDiffPlugin) Description() string {
	return "Plugin for comparing buffer contents and displaying differences"
}

func (p *BufferDiffPlugin) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
	fmt.Printf("[PLUGIN] Initialize called with host: %T\n", host)
	p.host = host
	fmt.Printf("[PLUGIN] Initialize completed, host stored: %v\n", p.host != nil)
	return nil
}

func (p *BufferDiffPlugin) Cleanup() error {
	return nil
}

func (p *BufferDiffPlugin) GetCommands() []pluginsdk.CommandSpec {
	fmt.Printf("[PLUGIN] GetCommands called, returning buffer diff commands\n")
	commands := []pluginsdk.CommandSpec{
		{
			Name:        "buffer-diff",
			Description: "Compare two buffers and show differences",
			Interactive: true,
			Handler:     "HandleBufferDiff",
			ArgPrompts:  []string{"Compare buffer: ", "With buffer: "},
		},
		{
			Name:        "buffer-diff-current",
			Description: "Compare current buffer with another buffer",
			Interactive: true,
			Handler:     "HandleBufferDiffCurrent",
			ArgPrompts:  []string{"Compare current buffer with: "},
		},
	}
	fmt.Printf("[PLUGIN] GetCommands returning %d commands: ", len(commands))
	for _, cmd := range commands {
		fmt.Printf("%s ", cmd.Name)
	}
	fmt.Printf("\n")
	return commands
}

func (p *BufferDiffPlugin) GetMajorModes() []pluginsdk.MajorModeSpec {
	return []pluginsdk.MajorModeSpec{}
}

func (p *BufferDiffPlugin) GetMinorModes() []pluginsdk.MinorModeSpec {
	return []pluginsdk.MinorModeSpec{}
}

func (p *BufferDiffPlugin) GetKeyBindings() []pluginsdk.KeyBindingSpec {
	return []pluginsdk.KeyBindingSpec{}
}

func (p *BufferDiffPlugin) HandleBufferDiff(buffer1Name, buffer2Name string) error {
	if p.host == nil {
		return fmt.Errorf("ERROR: host is nil")
	}

	fmt.Printf("[PLUGIN] HandleBufferDiff called with buffers: '%s' vs '%s'\n", buffer1Name, buffer2Name)

	// Find both buffers
	buffer1 := p.host.FindBuffer(buffer1Name)
	buffer2 := p.host.FindBuffer(buffer2Name)

	if buffer1 == nil {
		return fmt.Errorf("PLUGIN_MESSAGE:Buffer not found: %s", buffer1Name)
	}

	if buffer2 == nil {
		return fmt.Errorf("PLUGIN_MESSAGE:Buffer not found: %s", buffer2Name)
	}

	// Get buffer contents
	content1 := buffer1.Content()
	content2 := buffer2.Content()

	// Perform simple line-by-line diff
	diff := p.createSimpleDiff(buffer1Name, content1, buffer2Name, content2)

	// Create or find diff result buffer
	diffBufferName := fmt.Sprintf("*Diff: %s <-> %s*", buffer1Name, buffer2Name)
	diffBuffer := p.host.FindBuffer(diffBufferName)
	
	if diffBuffer == nil {
		diffBuffer = p.host.CreateBuffer(diffBufferName)
		if diffBuffer == nil {
			return fmt.Errorf("PLUGIN_MESSAGE:Failed to create diff buffer")
		}
	}

	// Clear and populate diff buffer
	diffBuffer.SetContent("")
	diffContent := strings.Join(diff, "\n")
	diffBuffer.SetContent(diffContent)

	// Switch to diff buffer
	err := p.host.SwitchToBuffer(diffBufferName)
	if err != nil {
		return fmt.Errorf("PLUGIN_MESSAGE:Failed to switch to diff buffer: %v", err)
	}

	return fmt.Errorf("PLUGIN_MESSAGE:Buffer diff completed: %d differences found", p.countDifferences(diff))
}

func (p *BufferDiffPlugin) HandleBufferDiffCurrent(otherBufferName string) error {
	if p.host == nil {
		return fmt.Errorf("ERROR: host is nil")
	}

	fmt.Printf("[PLUGIN] HandleBufferDiffCurrent called with buffer: '%s'\n", otherBufferName)

	// Get current buffer
	currentBuffer := p.host.GetCurrentBuffer()
	if currentBuffer == nil {
		return fmt.Errorf("PLUGIN_MESSAGE:No current buffer")
	}

	currentBufferName := currentBuffer.Name()
	return p.HandleBufferDiff(currentBufferName, otherBufferName)
}

func (p *BufferDiffPlugin) createSimpleDiff(name1, content1, name2, content2 string) []string {
	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")

	var result []string
	
	// Add header
	result = append(result, fmt.Sprintf("--- %s", name1))
	result = append(result, fmt.Sprintf("+++ %s", name2))
	result = append(result, "")

	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	// Simple line-by-line comparison
	for i := 0; i < maxLines; i++ {
		line1 := ""
		line2 := ""

		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 == line2 {
			result = append(result, fmt.Sprintf(" %s", line1))
		} else {
			if line1 != "" {
				result = append(result, fmt.Sprintf("-%s", line1))
			}
			if line2 != "" {
				result = append(result, fmt.Sprintf("+%s", line2))
			}
		}
	}

	return result
}

func (p *BufferDiffPlugin) countDifferences(diff []string) int {
	count := 0
	for _, line := range diff {
		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") {
			count++
		}
	}
	return count
}

// CommandPlugin インターフェース実装
func (p *BufferDiffPlugin) ExecuteCommand(name string, args ...interface{}) error {
	fmt.Printf("[PLUGIN] ExecuteCommand called: %s with %d args: %v\n", name, len(args), args)

	switch name {
	case "buffer-diff":
		if len(args) >= 2 {
			buffer1, ok1 := args[0].(string)
			buffer2, ok2 := args[1].(string)
			if ok1 && ok2 {
				return p.HandleBufferDiff(buffer1, buffer2)
			}
		}
		return fmt.Errorf("PLUGIN_MESSAGE:buffer-diff requires 2 buffer names")
	case "buffer-diff-current":
		if len(args) >= 1 {
			otherBuffer, ok := args[0].(string)
			if ok {
				return p.HandleBufferDiffCurrent(otherBuffer)
			}
		}
		return fmt.Errorf("PLUGIN_MESSAGE:buffer-diff-current requires 1 buffer name")
	default:
		return fmt.Errorf("unknown command: %s", name)
	}
}

func (p *BufferDiffPlugin) GetCompletions(command string, prefix string) []string {
	return []string{}
}

var pluginInstance = &BufferDiffPlugin{}

// RPCPlugin は標準的なgmacs RPCプラグイン実装
type RPCPlugin struct {
	Impl pluginsdk.Plugin
	broker *plugin.MuxBroker
}

func (p *RPCPlugin) Server(broker *plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl, broker: broker}, nil
}

func (p *RPCPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c, broker: b}, nil
}

// RPCServer はプラグイン側のRPCサーバー
type RPCServer struct {
	Impl   pluginsdk.Plugin
	broker *plugin.MuxBroker
}

// RPCClient はホスト側のRPCクライアント
type RPCClient struct {
	client *rpc.Client
	broker *plugin.MuxBroker
}

// Plugin インターフェースの実装（RPCClient）
func (c *RPCClient) Name() string {
	var resp string
	err := c.client.Call("Plugin.Name", interface{}(nil), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Version() string {
	var resp string
	err := c.client.Call("Plugin.Version", interface{}(nil), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Description() string {
	var resp string
	err := c.client.Call("Plugin.Description", interface{}(nil), &resp)
	if err != nil {
		return ""
	}
	return resp
}

func (c *RPCClient) Initialize(ctx context.Context, host pluginsdk.HostInterface) error {
	fmt.Printf("[RPC] Initialize called - setting up MuxBroker\n")
	
	// Start RPC server for HostInterface on client side
	hostBrokerID := c.broker.NextId()
	fmt.Printf("[RPC] Starting host RPC server with broker ID: %d\n", hostBrokerID)
	
	// Create a proper RPC server and register the Host service
	go func() {
		// Accept connection from plugin
		conn, err := c.broker.Accept(hostBrokerID)
		if err != nil {
			fmt.Printf("[RPC] Failed to accept connection on broker ID %d: %v\n", hostBrokerID, err)
			return
		}
		
		// Create RPC server and register Host service
		server := rpc.NewServer()
		err = server.RegisterName("Host", &RPCHostServer{Impl: host})
		if err != nil {
			fmt.Printf("[RPC] Failed to register Host service: %v\n", err)
			return
		}
		
		fmt.Printf("[RPC] Host service registered, serving RPC\n")
		server.ServeConn(conn)
	}()
	
	// Send the broker ID to plugin so it can connect back
	args := map[string]interface{}{
		"hostBrokerID": hostBrokerID,
	}
	
	fmt.Printf("[RPC] Calling Plugin.Initialize with args: %+v\n", args)
	var resp error
	err := c.client.Call("Plugin.Initialize", args, &resp)
	if err != nil {
		fmt.Printf("[RPC] Plugin.Initialize failed: %v\n", err)
	} else {
		fmt.Printf("[RPC] Plugin.Initialize succeeded\n")
	}
	return err
}

func (c *RPCClient) Cleanup() error {
	var resp error
	err := c.client.Call("Plugin.Cleanup", interface{}(nil), &resp)
	return err
}

func (c *RPCClient) GetCommands() []pluginsdk.CommandSpec {
	fmt.Printf("[RPC-Client] GetCommands called\n")
	var resp []pluginsdk.CommandSpec
	err := c.client.Call("Plugin.GetCommands", interface{}(nil), &resp)
	if err != nil {
		fmt.Printf("[RPC-Client] GetCommands RPC call failed: %v\n", err)
		return nil
	}
	fmt.Printf("[RPC-Client] GetCommands received %d commands: ", len(resp))
	for _, cmd := range resp {
		fmt.Printf("%s ", cmd.Name)
	}
	fmt.Printf("\n")
	return resp
}

func (c *RPCClient) GetMajorModes() []pluginsdk.MajorModeSpec {
	var resp []pluginsdk.MajorModeSpec
	err := c.client.Call("Plugin.GetMajorModes", interface{}(nil), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func (c *RPCClient) GetMinorModes() []pluginsdk.MinorModeSpec {
	var resp []pluginsdk.MinorModeSpec
	err := c.client.Call("Plugin.GetMinorModes", interface{}(nil), &resp)
	if err != nil {
		return nil
	}
	return resp
}

func (c *RPCClient) GetKeyBindings() []pluginsdk.KeyBindingSpec {
	var resp []pluginsdk.KeyBindingSpec
	err := c.client.Call("Plugin.GetKeyBindings", interface{}(nil), &resp)
	if err != nil {
		return nil
	}
	return resp
}

// RPCServer Plugin インターフェースの実装
func (s *RPCServer) Name(args interface{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *RPCServer) Version(args interface{}, resp *string) error {
	*resp = s.Impl.Version()
	return nil
}

func (s *RPCServer) Description(args interface{}, resp *string) error {
	*resp = s.Impl.Description()
	return nil
}

func (s *RPCServer) Initialize(args map[string]interface{}, resp *error) error {
	fmt.Printf("[RPC-Server] Initialize called with args: %+v\n", args)
	
	// Extract the host broker ID from args
	hostBrokerID, ok := args["hostBrokerID"].(uint32)
	if !ok {
		fmt.Printf("[RPC-Server] hostBrokerID not provided or wrong type\n")
		*resp = fmt.Errorf("hostBrokerID not provided")
		return nil
	}
	
	fmt.Printf("[RPC-Server] Connecting to host broker ID: %d\n", hostBrokerID)
	
	// Connect to the host's RPC server using MuxBroker
	conn, err := s.broker.Dial(hostBrokerID)
	if err != nil {
		fmt.Printf("[RPC-Server] Failed to connect to host broker: %v\n", err)
		*resp = fmt.Errorf("failed to connect to host broker: %v", err)
		return nil
	}
	
	fmt.Printf("[RPC-Server] Successfully connected to host broker\n")
	
	// Create RPC client for host interface
	hostClient := &RPCHostClient{client: rpc.NewClient(conn)}
	
	fmt.Printf("[RPC-Server] Created host RPC client, initializing plugin\n")
	
	// Initialize the plugin with the host interface
	*resp = s.Impl.Initialize(context.Background(), hostClient)
	if *resp != nil {
		fmt.Printf("[RPC-Server] Plugin initialization failed: %v\n", *resp)
	} else {
		fmt.Printf("[RPC-Server] Plugin initialization succeeded\n")
	}
	return nil
}

func (s *RPCServer) Cleanup(args interface{}, resp *error) error {
	*resp = s.Impl.Cleanup()
	return nil
}

func (s *RPCServer) GetCommands(args interface{}, resp *[]pluginsdk.CommandSpec) error {
	fmt.Printf("[RPC-Server] GetCommands called\n")
	commands := s.Impl.GetCommands()
	fmt.Printf("[RPC-Server] Got %d commands from Impl: ", len(commands))
	for _, cmd := range commands {
		fmt.Printf("%s ", cmd.Name)
	}
	fmt.Printf("\n")
	*resp = commands
	fmt.Printf("[RPC-Server] GetCommands setting resp to %d commands\n", len(*resp))
	return nil
}

func (s *RPCServer) GetMajorModes(args interface{}, resp *[]pluginsdk.MajorModeSpec) error {
	*resp = s.Impl.GetMajorModes()
	return nil
}

func (s *RPCServer) GetMinorModes(args interface{}, resp *[]pluginsdk.MinorModeSpec) error {
	*resp = s.Impl.GetMinorModes()
	return nil
}

func (s *RPCServer) GetKeyBindings(args interface{}, resp *[]pluginsdk.KeyBindingSpec) error {
	*resp = s.Impl.GetKeyBindings()
	return nil
}

func (s *RPCServer) ExecuteCommand(args map[string]interface{}, resp *string) error {
	name, _ := args["name"].(string)
	argsSlice, _ := args["args"].([]interface{})

	fmt.Printf("[RPC-Server] ExecuteCommand called: %s with args: %v\n", name, argsSlice)

	if cmdPlugin, ok := s.Impl.(interface{ ExecuteCommand(string, ...interface{}) error }); ok {
			err := cmdPlugin.ExecuteCommand(name, argsSlice...)
		if err != nil {
			fmt.Printf("[RPC-Server] ExecuteCommand failed: %v\n", err)
			*resp = err.Error()
		} else {
			*resp = ""
		}
	} else {
			*resp = "plugin does not support command execution"
	}
	return nil
}

// RPCHostClient はプラグイン側でホストの機能をRPC経由で呼び出すクライアント
type RPCHostClient struct {
	client *rpc.Client
}

// BufferInfo represents buffer state for RPC transmission
type BufferInfo struct {
	Name     string
	Content  string
	Position int
	IsDirty  bool
	Filename string
}

// RPCBufferProxy provides a client-side proxy for buffer operations via RPC
type RPCBufferProxy struct {
	client *rpc.Client
	info   BufferInfo
}

func (b *RPCBufferProxy) Name() string           { return b.info.Name }
func (b *RPCBufferProxy) Content() string        { return b.info.Content }
func (b *RPCBufferProxy) CursorPosition() int    { return b.info.Position }
func (b *RPCBufferProxy) IsDirty() bool          { return b.info.IsDirty }
func (b *RPCBufferProxy) Filename() string       { return b.info.Filename }

func (b *RPCBufferProxy) SetContent(content string) {
	b.info.Content = content
	// TODO: Implement RPC call to sync content to host
}

func (b *RPCBufferProxy) InsertAt(pos int, text string) {
	// TODO: Implement RPC call to insert text at position
}

func (b *RPCBufferProxy) DeleteRange(start, end int) {
	// TODO: Implement RPC call to delete text range
}

func (b *RPCBufferProxy) SetCursorPosition(pos int) {
	b.info.Position = pos
	// TODO: Implement RPC call to sync cursor position to host
}

func (b *RPCBufferProxy) MarkDirty() {
	b.info.IsDirty = true
	// TODO: Implement RPC call to mark buffer dirty on host
}

// HostInterface implementation for RPC client
func (h *RPCHostClient) GetCurrentBuffer() pluginsdk.BufferInterface {
	var resp BufferInfo
	err := h.client.Call("Host.GetCurrentBuffer", interface{}(nil), &resp)
	if err != nil {
		fmt.Printf("[RPC] GetCurrentBuffer call failed: %v\n", err)
		return nil
	}
	
	return &RPCBufferProxy{
		client: h.client,
		info:   resp,
	}
}

func (h *RPCHostClient) GetCurrentWindow() pluginsdk.WindowInterface {
	// TODO: Implement RPC call to host
	return nil
}

func (h *RPCHostClient) SetStatus(message string) {
	var resp error
	h.client.Call("Host.SetStatus", message, &resp)
}

func (h *RPCHostClient) ShowMessage(message string) {
	var resp error
	h.client.Call("Host.ShowMessage", message, &resp)
}

func (h *RPCHostClient) ExecuteCommand(name string, args ...interface{}) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("ExecuteCommand not implemented in RPC client")
}

func (h *RPCHostClient) SetMajorMode(bufferName, modeName string) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("SetMajorMode not implemented in RPC client")
}

func (h *RPCHostClient) ToggleMinorMode(bufferName, modeName string) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("ToggleMinorMode not implemented in RPC client")
}

func (h *RPCHostClient) AddHook(event string, handler func(...interface{}) error) {
	// TODO: Implement RPC call to host
}

func (h *RPCHostClient) TriggerHook(event string, args ...interface{}) {
	// TODO: Implement RPC call to host
}

func (h *RPCHostClient) CreateBuffer(name string) pluginsdk.BufferInterface {
	fmt.Printf("[RPC] CreateBuffer called with name: %s\n", name)
	var resp BufferInfo
	err := h.client.Call("Host.CreateBuffer", name, &resp)
	if err != nil {
		fmt.Printf("[RPC] CreateBuffer call failed: %v\n", err)
		return nil
	}
	
	fmt.Printf("[RPC] CreateBuffer succeeded: %+v\n", resp)
	return &RPCBufferProxy{
		client: h.client,
		info:   resp,
	}
}

func (h *RPCHostClient) FindBuffer(name string) pluginsdk.BufferInterface {
	fmt.Printf("[RPC] FindBuffer called with name: %s\n", name)
	var resp BufferInfo
	err := h.client.Call("Host.FindBuffer", name, &resp)
	if err != nil {
		fmt.Printf("[RPC] FindBuffer call failed: %v\n", err)
		return nil
	}
	
	// Check if buffer was found (empty name means not found)
	if resp.Name == "" {
		fmt.Printf("[RPC] FindBuffer: buffer '%s' not found\n", name)
		return nil
	}
	
	fmt.Printf("[RPC] FindBuffer succeeded: %+v\n", resp)
	return &RPCBufferProxy{
		client: h.client,
		info:   resp,
	}
}

func (h *RPCHostClient) SwitchToBuffer(name string) error {
	var resp error
	err := h.client.Call("Host.SwitchToBuffer", name, &resp)
	if err != nil {
		return fmt.Errorf("RPC call failed: %v", err)
	}
	return resp
}

func (h *RPCHostClient) OpenFile(path string) error {
	fmt.Printf("[RPC] OpenFile called with path: %s\n", path)
	var resp error
	err := h.client.Call("Host.OpenFile", path, &resp)
	if err != nil {
		fmt.Printf("[RPC] OpenFile RPC call failed: %v\n", err)
		return fmt.Errorf("RPC call failed: %v", err)
	}
	if resp != nil {
		fmt.Printf("[RPC] OpenFile failed on host: %v\n", resp)
	} else {
		fmt.Printf("[RPC] OpenFile succeeded on host\n")
	}
	return resp
}

func (h *RPCHostClient) SaveBuffer(bufferName string) error {
	fmt.Printf("[RPC] SaveBuffer called with buffer: %s\n", bufferName)
	var resp error
	err := h.client.Call("Host.SaveBuffer", bufferName, &resp)
	if err != nil {
		fmt.Printf("[RPC] SaveBuffer RPC call failed: %v\n", err)
		return fmt.Errorf("RPC call failed: %v", err)
	}
	if resp != nil {
		fmt.Printf("[RPC] SaveBuffer failed on host: %v\n", resp)
	} else {
		fmt.Printf("[RPC] SaveBuffer succeeded on host\n")
	}
	return resp
}

func (h *RPCHostClient) GetOption(name string) (interface{}, error) {
	// TODO: Implement RPC call to host
	return nil, fmt.Errorf("GetOption not implemented in RPC client")
}

func (h *RPCHostClient) SetOption(name string, value interface{}) error {
	// TODO: Implement RPC call to host
	return fmt.Errorf("SetOption not implemented in RPC client")
}

// RPCHostServer はgmacs側でホスト機能をRPC経由で提供するサーバー
type RPCHostServer struct {
	Impl pluginsdk.HostInterface
}

func (h *RPCHostServer) SetStatus(message string, resp *error) error {
	h.Impl.SetStatus(message)
	*resp = nil
	return nil
}

// CreateBuffer handles RPC calls from plugins to create buffers
func (h *RPCHostServer) CreateBuffer(name string, resp *BufferInfo) error {
	buffer := h.Impl.CreateBuffer(name)
	if buffer == nil {
		*resp = BufferInfo{}
		return fmt.Errorf("failed to create buffer")
	}
	
	// Return buffer information via RPC
	*resp = BufferInfo{
		Name:     buffer.Name(),
		Content:  buffer.Content(),
		Position: buffer.CursorPosition(),
		IsDirty:  buffer.IsDirty(),
		Filename: buffer.Filename(),
	}
	return nil
}

// FindBuffer handles RPC calls from plugins to find buffers
func (h *RPCHostServer) FindBuffer(name string, resp *BufferInfo) error {
	fmt.Printf("[RPC-Host] FindBuffer called with name: %s\n", name)
	buffer := h.Impl.FindBuffer(name)
	if buffer == nil {
		fmt.Printf("[RPC-Host] FindBuffer: buffer '%s' not found\n", name)
		*resp = BufferInfo{} // Empty response indicates not found
		return nil
	}
	
	fmt.Printf("[RPC-Host] FindBuffer: found buffer '%s'\n", buffer.Name())
	
	// Return buffer information via RPC
	*resp = BufferInfo{
		Name:     buffer.Name(),
		Content:  buffer.Content(),
		Position: buffer.CursorPosition(),
		IsDirty:  buffer.IsDirty(),
		Filename: buffer.Filename(),
	}
	
	fmt.Printf("[RPC-Host] FindBuffer: returning buffer info: %+v\n", *resp)
	return nil
}

// GetCurrentBuffer handles RPC calls from plugins to get current buffer
func (h *RPCHostServer) GetCurrentBuffer(args interface{}, resp *BufferInfo) error {
	fmt.Printf("[RPC-Host] GetCurrentBuffer called\n")
	buffer := h.Impl.GetCurrentBuffer()
	if buffer == nil {
		fmt.Printf("[RPC-Host] GetCurrentBuffer: no current buffer found\n")
		*resp = BufferInfo{}
		return fmt.Errorf("no current buffer")
	}
	
	fmt.Printf("[RPC-Host] GetCurrentBuffer: found buffer '%s'\n", buffer.Name())
	
	// Return buffer information via RPC
	*resp = BufferInfo{
		Name:     buffer.Name(),
		Content:  buffer.Content(),
		Position: buffer.CursorPosition(),
		IsDirty:  buffer.IsDirty(),
		Filename: buffer.Filename(),
	}
	
	fmt.Printf("[RPC-Host] GetCurrentBuffer: returning buffer info: %+v\n", *resp)
	return nil
}

// SwitchToBuffer handles RPC calls from plugins to switch buffers
func (h *RPCHostServer) SwitchToBuffer(name string, resp *error) error {
	*resp = h.Impl.SwitchToBuffer(name)
	return nil
}

// OpenFile handles RPC calls from plugins to open files
func (h *RPCHostServer) OpenFile(path string, resp *error) error {
	fmt.Printf("[RPC-Host] OpenFile called with path: %s\n", path)
	*resp = h.Impl.OpenFile(path)
	if *resp != nil {
		fmt.Printf("[RPC-Host] OpenFile failed: %v\n", *resp)
	} else {
		fmt.Printf("[RPC-Host] OpenFile succeeded for: %s\n", path)
	}
	return nil
}

// SaveBuffer handles RPC calls from plugins to save buffers
func (h *RPCHostServer) SaveBuffer(bufferName string, resp *error) error {
	fmt.Printf("[RPC-Host] SaveBuffer called with buffer: %s\n", bufferName)
	*resp = h.Impl.SaveBuffer(bufferName)
	if *resp != nil {
		fmt.Printf("[RPC-Host] SaveBuffer failed: %v\n", *resp)
	} else {
		fmt.Printf("[RPC-Host] SaveBuffer succeeded for: %s\n", bufferName)
	}
	return nil
}