package main

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
	in := strings.NewReader("a,1,A\nb,2,B\nc,3,C\n")
	reader = bufio.NewScanner(in)
	buf := &bytes.Buffer{}
	writer = bufio.NewWriter(buf)
	c, err := NewCutter(",", "1,3", false, false, 0)
	require.NoError(t, err)
	c.extract()
	writer.Flush()
	require.Equal(t, "a,A\nb,B\nc,C\n", buf.String())
}
