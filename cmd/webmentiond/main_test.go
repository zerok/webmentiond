package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandConstruction(t *testing.T) {
	cmd := newRootCmd()

	t.Run("noop", func(t *testing.T) {
		c := cmd.Cmd()
		err := c.Execute()
		require.NoError(t, err)
	})
}
