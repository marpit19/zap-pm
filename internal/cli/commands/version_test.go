package commands

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.Execute()
	
	out := b.String()
	assert.Contains(t, out, "Zap Package Manager v")
}