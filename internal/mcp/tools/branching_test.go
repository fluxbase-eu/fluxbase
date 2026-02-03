package tools

import (
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ListBranchesTool Tests
// =============================================================================

func TestListBranchesTool_Name(t *testing.T) {
	tool := NewListBranchesTool(nil)
	assert.Equal(t, "list_branches", tool.Name())
}

func TestListBranchesTool_Description(t *testing.T) {
	tool := NewListBranchesTool(nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "List")
	assert.Contains(t, desc, "branches")
}

func TestListBranchesTool_InputSchema(t *testing.T) {
	tool := NewListBranchesTool(nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has status property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		statusProp, ok := properties["status"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", statusProp["type"])

		enum, ok := statusProp["enum"].([]string)
		require.True(t, ok)
		assert.Contains(t, enum, "creating")
		assert.Contains(t, enum, "ready")
		assert.Contains(t, enum, "error")
	})

	t.Run("has type property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		typeProp, ok := properties["type"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", typeProp["type"])

		enum, ok := typeProp["enum"].([]string)
		require.True(t, ok)
		assert.Contains(t, enum, "main")
		assert.Contains(t, enum, "preview")
		assert.Contains(t, enum, "persistent")
	})

	t.Run("has limit property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		limitProp, ok := properties["limit"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "integer", limitProp["type"])
		assert.Equal(t, 50, limitProp["default"])
	})

	t.Run("has offset property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		offsetProp, ok := properties["offset"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "integer", offsetProp["type"])
		assert.Equal(t, 0, offsetProp["default"])
	})
}

func TestListBranchesTool_RequiredScopes(t *testing.T) {
	tool := NewListBranchesTool(nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchRead)
}

// =============================================================================
// GetBranchTool Tests
// =============================================================================

func TestGetBranchTool_Name(t *testing.T) {
	tool := NewGetBranchTool(nil)
	assert.Equal(t, "get_branch", tool.Name())
}

func TestGetBranchTool_Description(t *testing.T) {
	tool := NewGetBranchTool(nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "Get")
	assert.Contains(t, desc, "branch")
}

func TestGetBranchTool_InputSchema(t *testing.T) {
	tool := NewGetBranchTool(nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has branch_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		branchIDProp, ok := properties["branch_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", branchIDProp["type"])
	})

	t.Run("has slug property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		slugProp, ok := properties["slug"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", slugProp["type"])
	})
}

func TestGetBranchTool_RequiredScopes(t *testing.T) {
	tool := NewGetBranchTool(nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchRead)
}

// =============================================================================
// CreateBranchTool Tests
// =============================================================================

func TestCreateBranchTool_Name(t *testing.T) {
	tool := NewCreateBranchTool(nil)
	assert.Equal(t, "create_branch", tool.Name())
}

func TestCreateBranchTool_Description(t *testing.T) {
	tool := NewCreateBranchTool(nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "Create")
	assert.Contains(t, desc, "branch")
}

func TestCreateBranchTool_InputSchema(t *testing.T) {
	tool := NewCreateBranchTool(nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has name property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		nameProp, ok := properties["name"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", nameProp["type"])
	})

	t.Run("has parent_branch_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		parentProp, ok := properties["parent_branch_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", parentProp["type"])
	})

	t.Run("has data_clone_mode property with enum", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		dcmProp, ok := properties["data_clone_mode"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", dcmProp["type"])

		enum, ok := dcmProp["enum"].([]string)
		require.True(t, ok)
		assert.Contains(t, enum, "schema_only")
		assert.Contains(t, enum, "full_clone")
		assert.Contains(t, enum, "seed_data")
	})

	t.Run("has type property with enum", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		typeProp, ok := properties["type"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", typeProp["type"])

		enum, ok := typeProp["enum"].([]string)
		require.True(t, ok)
		assert.Contains(t, enum, "preview")
		assert.Contains(t, enum, "persistent")
	})

	t.Run("has expires_at property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		expiresProp, ok := properties["expires_at"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", expiresProp["type"])
	})

	t.Run("requires name", func(t *testing.T) {
		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "name")
	})
}

func TestCreateBranchTool_RequiredScopes(t *testing.T) {
	tool := NewCreateBranchTool(nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchWrite)
}

// =============================================================================
// DeleteBranchTool Tests
// =============================================================================

func TestDeleteBranchTool_Name(t *testing.T) {
	tool := NewDeleteBranchTool(nil, nil)
	assert.Equal(t, "delete_branch", tool.Name())
}

func TestDeleteBranchTool_Description(t *testing.T) {
	tool := NewDeleteBranchTool(nil, nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "Delete")
	assert.Contains(t, desc, "branch")
}

func TestDeleteBranchTool_InputSchema(t *testing.T) {
	tool := NewDeleteBranchTool(nil, nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has branch_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		branchIDProp, ok := properties["branch_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", branchIDProp["type"])
	})

	t.Run("has slug property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		slugProp, ok := properties["slug"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", slugProp["type"])
	})
}

func TestDeleteBranchTool_RequiredScopes(t *testing.T) {
	tool := NewDeleteBranchTool(nil, nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchWrite)
}

// =============================================================================
// ResetBranchTool Tests
// =============================================================================

func TestResetBranchTool_Name(t *testing.T) {
	tool := NewResetBranchTool(nil, nil)
	assert.Equal(t, "reset_branch", tool.Name())
}

func TestResetBranchTool_Description(t *testing.T) {
	tool := NewResetBranchTool(nil, nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "Reset")
	assert.Contains(t, desc, "branch")
}

func TestResetBranchTool_InputSchema(t *testing.T) {
	tool := NewResetBranchTool(nil, nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has branch_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		branchIDProp, ok := properties["branch_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", branchIDProp["type"])
	})

	t.Run("has slug property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		slugProp, ok := properties["slug"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", slugProp["type"])
	})
}

func TestResetBranchTool_RequiredScopes(t *testing.T) {
	tool := NewResetBranchTool(nil, nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchWrite)
}

// =============================================================================
// GrantBranchAccessTool Tests
// =============================================================================

func TestGrantBranchAccessTool_Name(t *testing.T) {
	tool := NewGrantBranchAccessTool(nil)
	assert.Equal(t, "grant_branch_access", tool.Name())
}

func TestGrantBranchAccessTool_Description(t *testing.T) {
	tool := NewGrantBranchAccessTool(nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "Grant")
	assert.Contains(t, desc, "access")
}

func TestGrantBranchAccessTool_InputSchema(t *testing.T) {
	tool := NewGrantBranchAccessTool(nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has branch_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		branchIDProp, ok := properties["branch_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", branchIDProp["type"])
	})

	t.Run("has user_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		userIDProp, ok := properties["user_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", userIDProp["type"])
	})

	t.Run("has access_level property with enum", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		accessLevelProp, ok := properties["access_level"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", accessLevelProp["type"])

		enum, ok := accessLevelProp["enum"].([]string)
		require.True(t, ok)
		assert.Contains(t, enum, "read")
		assert.Contains(t, enum, "write")
		assert.Contains(t, enum, "admin")
	})

	t.Run("requires all three fields", func(t *testing.T) {
		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "branch_id")
		assert.Contains(t, required, "user_id")
		assert.Contains(t, required, "access_level")
	})
}

func TestGrantBranchAccessTool_RequiredScopes(t *testing.T) {
	tool := NewGrantBranchAccessTool(nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchAccess)
}

// =============================================================================
// RevokeBranchAccessTool Tests
// =============================================================================

func TestRevokeBranchAccessTool_Name(t *testing.T) {
	tool := NewRevokeBranchAccessTool(nil)
	assert.Equal(t, "revoke_branch_access", tool.Name())
}

func TestRevokeBranchAccessTool_Description(t *testing.T) {
	tool := NewRevokeBranchAccessTool(nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "Revoke")
	assert.Contains(t, desc, "access")
}

func TestRevokeBranchAccessTool_InputSchema(t *testing.T) {
	tool := NewRevokeBranchAccessTool(nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has branch_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		branchIDProp, ok := properties["branch_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", branchIDProp["type"])
	})

	t.Run("has user_id property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		userIDProp, ok := properties["user_id"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", userIDProp["type"])
	})

	t.Run("requires branch_id and user_id", func(t *testing.T) {
		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "branch_id")
		assert.Contains(t, required, "user_id")
	})
}

func TestRevokeBranchAccessTool_RequiredScopes(t *testing.T) {
	tool := NewRevokeBranchAccessTool(nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchAccess)
}

// =============================================================================
// GetActiveBranchTool Tests
// =============================================================================

func TestGetActiveBranchTool_Name(t *testing.T) {
	tool := NewGetActiveBranchTool(nil)
	assert.Equal(t, "get_active_branch", tool.Name())
}

func TestGetActiveBranchTool_Description(t *testing.T) {
	tool := NewGetActiveBranchTool(nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "active")
	assert.Contains(t, desc, "branch")
}

func TestGetActiveBranchTool_InputSchema(t *testing.T) {
	tool := NewGetActiveBranchTool(nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has empty properties", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)
		assert.Empty(t, properties)
	})
}

func TestGetActiveBranchTool_RequiredScopes(t *testing.T) {
	tool := NewGetActiveBranchTool(nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchRead)
}

// =============================================================================
// SetActiveBranchTool Tests
// =============================================================================

func TestSetActiveBranchTool_Name(t *testing.T) {
	tool := NewSetActiveBranchTool(nil, nil)
	assert.Equal(t, "set_active_branch", tool.Name())
}

func TestSetActiveBranchTool_Description(t *testing.T) {
	tool := NewSetActiveBranchTool(nil, nil)
	desc := tool.Description()

	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "Set")
	assert.Contains(t, desc, "active")
}

func TestSetActiveBranchTool_InputSchema(t *testing.T) {
	tool := NewSetActiveBranchTool(nil, nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has branch property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		branchProp, ok := properties["branch"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", branchProp["type"])
	})

	t.Run("requires branch", func(t *testing.T) {
		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "branch")
	})
}

func TestSetActiveBranchTool_RequiredScopes(t *testing.T) {
	tool := NewSetActiveBranchTool(nil, nil)
	scopes := tool.RequiredScopes()

	assert.Contains(t, scopes, mcp.ScopeBranchWrite)
}

// =============================================================================
// Tool Struct Tests
// =============================================================================

func TestBranchingTool_Structs(t *testing.T) {
	t.Run("ListBranchesTool stores storage", func(t *testing.T) {
		tool := &ListBranchesTool{storage: nil}
		assert.Nil(t, tool.storage)
	})

	t.Run("GetBranchTool stores storage", func(t *testing.T) {
		tool := &GetBranchTool{storage: nil}
		assert.Nil(t, tool.storage)
	})

	t.Run("CreateBranchTool stores manager", func(t *testing.T) {
		tool := &CreateBranchTool{manager: nil}
		assert.Nil(t, tool.manager)
	})

	t.Run("DeleteBranchTool stores manager and storage", func(t *testing.T) {
		tool := &DeleteBranchTool{manager: nil, storage: nil}
		assert.Nil(t, tool.manager)
		assert.Nil(t, tool.storage)
	})

	t.Run("ResetBranchTool stores manager and storage", func(t *testing.T) {
		tool := &ResetBranchTool{manager: nil, storage: nil}
		assert.Nil(t, tool.manager)
		assert.Nil(t, tool.storage)
	})

	t.Run("GrantBranchAccessTool stores storage", func(t *testing.T) {
		tool := &GrantBranchAccessTool{storage: nil}
		assert.Nil(t, tool.storage)
	})

	t.Run("RevokeBranchAccessTool stores storage", func(t *testing.T) {
		tool := &RevokeBranchAccessTool{storage: nil}
		assert.Nil(t, tool.storage)
	})

	t.Run("GetActiveBranchTool stores router", func(t *testing.T) {
		tool := &GetActiveBranchTool{router: nil}
		assert.Nil(t, tool.router)
	})

	t.Run("SetActiveBranchTool stores router and storage", func(t *testing.T) {
		tool := &SetActiveBranchTool{router: nil, storage: nil}
		assert.Nil(t, tool.router)
		assert.Nil(t, tool.storage)
	})
}

// =============================================================================
// ToolHandler Interface Compliance Tests
// =============================================================================

func TestBranchingTools_ImplementsToolHandler(t *testing.T) {
	t.Run("ListBranchesTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &ListBranchesTool{}
	})

	t.Run("GetBranchTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &GetBranchTool{}
	})

	t.Run("CreateBranchTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &CreateBranchTool{}
	})

	t.Run("DeleteBranchTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &DeleteBranchTool{}
	})

	t.Run("ResetBranchTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &ResetBranchTool{}
	})

	t.Run("GrantBranchAccessTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &GrantBranchAccessTool{}
	})

	t.Run("RevokeBranchAccessTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &RevokeBranchAccessTool{}
	})

	t.Run("GetActiveBranchTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &GetActiveBranchTool{}
	})

	t.Run("SetActiveBranchTool implements ToolHandler", func(t *testing.T) {
		var _ ToolHandler = &SetActiveBranchTool{}
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkListBranchesTool_InputSchema(b *testing.B) {
	tool := NewListBranchesTool(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.InputSchema()
	}
}

func BenchmarkCreateBranchTool_InputSchema(b *testing.B) {
	tool := NewCreateBranchTool(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.InputSchema()
	}
}

func BenchmarkGrantBranchAccessTool_InputSchema(b *testing.B) {
	tool := NewGrantBranchAccessTool(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.InputSchema()
	}
}
