package imports_test

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/fellou89/gogen/imports"
	"github.com/stretchr/testify/require"
)

// TestDevendorizeImportPaths checks if vendored
// import paths are devendorized correctly.
func TestDevendorizeImportPaths(t *testing.T) {
	i := imports.New("github.com/fellou89/gogen/imports")
	pkg := types.NewPackage("github.com/fellou89/gogen/vendor/github.com/stretchr/testify/mock", "mock")
	named := types.NewNamed(types.NewTypeName(token.Pos(0), pkg, "", &types.Array{}), &types.Array{}, nil)
	i.AddImportsFrom(named)
	require.Equal(t, map[string]string{"github.com/stretchr/testify/mock": "mock"}, i.Imports())
}
