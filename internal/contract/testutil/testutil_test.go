package testutil_test

import (
	"testing"

	"github.com/alfred/alfred/internal/contract"
	"github.com/alfred/alfred/internal/contract/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewThread(t *testing.T) {
	t.Parallel()
	th := testutil.NewThread("thr-1", "smoke")
	require.NotNil(t, th)
	assert.Equal(t, "thr-1", string(th.ID))
	assert.Equal(t, "smoke", th.Title)
	assert.Equal(t, contract.ThreadStatusIdle, th.Status)
	assert.False(t, th.CreatedAt.IsZero())
}

func TestNewTurn(t *testing.T) {
	t.Parallel()
	tr := testutil.NewTurn("turn-1", "thr-1", contract.TurnStatusCompleted)
	require.NotNil(t, tr)
	assert.Equal(t, "turn-1", string(tr.ID))
	assert.Equal(t, "thr-1", string(tr.ThreadID))
	assert.Equal(t, contract.TurnStatusCompleted, tr.Status)
	assert.Empty(t, tr.Items)
}

func TestNewUserMessageItem(t *testing.T) {
	t.Parallel()
	it := testutil.NewUserMessageItem("it-1", "hi")
	assert.Equal(t, contract.ItemKindUserMessage, it.Kind)
	require.NotNil(t, it.Text)
	assert.Equal(t, "hi", *it.Text)
}

func TestNewAssistantTextItem(t *testing.T) {
	t.Parallel()
	it := testutil.NewAssistantTextItem("it-2", "hello")
	assert.Equal(t, contract.ItemKindAssistantText, it.Kind)
	require.NotNil(t, it.Text)
	assert.Equal(t, "hello", *it.Text)
}

func TestNewToolResultItem(t *testing.T) {
	t.Parallel()
	it := testutil.NewToolResultItem("it-3", "echo", []byte(`{"ok":true}`), []byte(`{}`))
	require.NotNil(t, it.ToolCall)
	assert.Equal(t, "echo", it.ToolCall.Name)
}

func TestNewUserInput(t *testing.T) {
	t.Parallel()
	in := testutil.NewUserInput("hello")
	assert.Equal(t, "hello", in.Text)
	assert.Empty(t, in.DisplayText)
	assert.Empty(t, in.Attachments)
}

func TestTestutilHelpersComposeIntoTurn(t *testing.T) {
	t.Parallel()
	tr := testutil.NewTurn("turn-1", "thr-1", contract.TurnStatusRunning)
	tr.Items = append(tr.Items,
		testutil.NewUserMessageItem("u", "hi"),
		testutil.NewAssistantTextItem("a", "hello"),
	)
	assert.Len(t, tr.Items, 2)
	assert.Equal(t, contract.ItemKindUserMessage, tr.Items[0].Kind)
	assert.Equal(t, contract.ItemKindAssistantText, tr.Items[1].Kind)
}
