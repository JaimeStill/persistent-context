#!/bin/bash

# Test MCP tools functionality
MCP_BIN="./bin/persistent-context-mcp"

echo "Testing MCP Tools..."

# Test with a single session
echo "Running MCP session test..."
(
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"tools": {}}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}}'
sleep 0.1
echo '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}'
sleep 0.1
echo '{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "get_stats"}}'
sleep 0.1
echo '{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "capture_memory", "arguments": {"source": "test", "content": "MCP test memory"}}}'
sleep 0.1
echo '{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "get_memories", "arguments": {"limit": 5}}}'
sleep 0.1
echo '{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "search_memories", "arguments": {"content": "test"}}}'
sleep 0.1
echo '{"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": {"name": "trigger_consolidation"}}'
) | $MCP_BIN --stdio

echo "MCP testing complete!"