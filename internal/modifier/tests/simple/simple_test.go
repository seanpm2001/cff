package simple

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlow(t *testing.T) {
	iRes, sRes, err := Flow()
	assert.NoError(t, err)
	assert.Equal(t, iRes, 1)
	assert.Equal(t, sRes, "non-zero")
}

func TestModifyVarInScope(t *testing.T) {
	res, side, err := ModifyVarInScope()
	assert.NoError(t, err)
	assert.Equal(t, res, true)
	assert.Equal(t, side, []int{1, 2, 3})
}

func TestExternal(t *testing.T) {
	res, err := External()
	assert.NoError(t, err)
	assert.Equal(t, res, true)
}
