package agent

import (
	"encoding/json"
	"sort"

	"reasonix/internal/plugin"
	"reasonix/internal/tool"
)

func (t *UseCapabilityTool) mcpArgumentKeys(id string) ([]string, []string) {
	server, raw, err := parseMCPCapabilityID(id)
	if err != nil || !t.serverEnabled(server) {
		return nil, nil
	}
	if schema := t.registeredMCPSchema(server, raw); len(schema) > 0 {
		return argumentKeysFromSchema(schema)
	}
	if t.state != nil {
		if schema := cachedToolSchema(t.state.snapshotLiveTools()[server], raw); len(schema) > 0 {
			return argumentKeysFromSchema(schema)
		}
	}
	if t.runtime != nil {
		_, cached, keyOK, _, _ := t.runtime.CapabilityCatalogState()
		if keyOK[server] {
			return argumentKeysFromSchema(cachedToolSchema(cached[server], raw))
		}
	}
	return nil, nil
}

func (t *UseCapabilityTool) registeredMCPSchema(server, raw string) json.RawMessage {
	if t.registry == nil {
		return nil
	}
	target, ok := t.registry.Get(plugin.ModelToolName(server, raw))
	if !ok {
		return nil
	}
	metadata, ok := target.(tool.MCPMetadata)
	if !ok || metadata.MCPServerName() != server || metadata.MCPRawToolName() != raw {
		return nil
	}
	return target.Schema()
}

func cachedToolSchema(tools []plugin.CachedTool, raw string) json.RawMessage {
	for _, candidate := range tools {
		if candidate.Name == raw {
			return candidate.Schema
		}
	}
	return nil
}

func argumentKeysFromSchema(raw json.RawMessage) ([]string, []string) {
	if len(raw) == 0 {
		return nil, nil
	}
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if json.Unmarshal(raw, &schema) != nil || len(schema.Properties) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(schema.Properties))
	for key := range schema.Properties {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	required := make([]string, 0, len(schema.Required))
	seen := make(map[string]bool, len(schema.Required))
	for _, key := range schema.Required {
		if _, ok := schema.Properties[key]; ok && !seen[key] {
			seen[key] = true
			required = append(required, key)
		}
	}
	sort.Strings(required)
	return keys, required
}
