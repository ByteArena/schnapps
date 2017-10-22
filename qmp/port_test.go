package qmp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetNextPort(t *testing.T) {
	assert.Equal(t, GetNextPort(), 44400)
	assert.Equal(t, GetNextPort(), 44401)
}

func TestGetNextPortReset(t *testing.T) {
	inc = MAX

	assert.Equal(t, GetNextPort(), 44499)
	assert.Equal(t, GetNextPort(), 44400)
}
