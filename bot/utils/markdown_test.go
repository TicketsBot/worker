package utils

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNoEscape(t *testing.T) {
	input := "hello world"
	require.Equal(t, input, EscapeMarkdown(input))
}

func TestEscapeBold(t *testing.T) {
	input := "hello **world**"
	require.Equal(t, "hello \\*\\*world\\*\\*", EscapeMarkdown(input))
}

func TestEscapeMulti(t *testing.T) {
	input := "hello __**world**__"
	require.Equal(t, "hello \\_\\_\\*\\*world\\*\\*\\_\\_", EscapeMarkdown(input))
}

func TestEscapeLink(t *testing.T) {
	input := "hello https://google.com/some_path_here **hello world**"
	expected := "hello https://google.com/some_path_here \\*\\*hello world\\*\\*"
	require.Equal(t, expected, EscapeMarkdown(input))
}

func TestHttpsIncomplete(t *testing.T) {
	input := "hello https:/"
	require.Equal(t, input, EscapeMarkdown(input))
}

func TestHttpIncomplete(t *testing.T) {
	input := "hello http:/"
	require.Equal(t, input, EscapeMarkdown(input))
}
