package runtime

import (
	"testing"

	"github.com/sachncs/alfred/internal/contract"
)

func TestAlfredRuntimePublishEventCapsReplay(t *testing.T) {
	rt := newTestAlfredRuntime(t)
	threadID := contract.ThreadID("thr-cap")
	for i := 0; i < maxReplayPerThread*4; i++ {
		rt.PublishEvent(threadID, SSEEvent{ID: "e", Event: "x", Data: i})
	}
	rt.replayMu.RLock()
	buf := rt.replayBuf[threadID]
	rt.replayMu.RUnlock()
	if len(buf) != maxReplayPerThread {
		t.Errorf("replay buffer length = %d, want %d", len(buf), maxReplayPerThread)
	}
	first, ok := buf[0].Data.(int)
	if !ok {
		t.Fatalf("oldest event data not int: %T", buf[0].Data)
	}
	if first != maxReplayPerThread*3 {
		t.Errorf("oldest event data = %d, want %d", first, maxReplayPerThread*3)
	}
}
