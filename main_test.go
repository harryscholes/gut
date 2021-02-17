package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {

	t.Run("select columns", func(t *testing.T) {
		in := strings.NewReader("a,1,A\nb,2,B\nc,3,C\n")
		reader = bufio.NewScanner(in)
		buf := &bytes.Buffer{}
		writer = bufio.NewWriter(buf)
		c, err := NewCutter(",", "1,3", false, false, 0)
		require.NoError(t, err)
		c.extract()
		writer.Flush()
		require.Equal(t, "a,A\nb,B\nc,C\n", buf.String())
	})

	t.Run("reorder columns", func(t *testing.T) {
		in := strings.NewReader("a,1,A\nb,2,B\nc,3,C\n")
		reader = bufio.NewScanner(in)
		buf := &bytes.Buffer{}
		writer = bufio.NewWriter(buf)
		c, err := NewCutter(",", "2,3,1", false, false, 0)
		require.NoError(t, err)
		c.extract()
		writer.Flush()
		require.Equal(t, "1,A,a\n2,B,b\n3,C,c\n", buf.String())
	})
}
