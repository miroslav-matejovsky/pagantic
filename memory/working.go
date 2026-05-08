package memory

import "github.com/miroslav-matejovsky/pagantic/core"

// WorkingMemory holds transient context for current execution step.
// Reset between steps by orchestration layer.
type WorkingMemory struct {
	StepResults map[string]any
	Context     []core.Message
}

// NewWorkingMemory creates empty working memory.
func NewWorkingMemory() *WorkingMemory {
	return &WorkingMemory{StepResults: make(map[string]any)}
}

// Reset clears transient state.
func (wm *WorkingMemory) Reset() {
	wm.StepResults = make(map[string]any)
	wm.Context = nil
}

// SetResult stores result for key.
func (wm *WorkingMemory) SetResult(key string, value any) {
	if wm.StepResults == nil {
		wm.StepResults = make(map[string]any)
	}

	wm.StepResults[key] = value
}

// GetResult loads result for key.
func (wm *WorkingMemory) GetResult(key string) (any, bool) {
	value, ok := wm.StepResults[key]
	return value, ok
}
