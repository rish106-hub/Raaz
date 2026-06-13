package store

import "time"

var seedPrompts = []Prompt{
	{ID: "p01", Text: "What's one thing you're secretly proud of?", Category: "self"},
	{ID: "p02", Text: "If you could travel anywhere tomorrow, where would you go and why?", Category: "travel"},
	{ID: "p03", Text: "What's a belief you hold that most people around you don't share?", Category: "philosophy"},
	{ID: "p04", Text: "What's the last thing that genuinely made you laugh?", Category: "lighthearted"},
	{ID: "p05", Text: "What's something you've changed your mind about in the last year?", Category: "growth"},
	{ID: "p06", Text: "Describe your perfect Saturday from morning to night.", Category: "lifestyle"},
	{ID: "p07", Text: "What's a skill you've always wanted to learn but never started?", Category: "aspiration"},
	{ID: "p08", Text: "What does home mean to you?", Category: "meaning"},
	{ID: "p09", Text: "What's the most interesting conversation you've had recently?", Category: "connection"},
	{ID: "p10", Text: "If your life had a soundtrack, what song would be playing right now?", Category: "identity"},
}

// MemoryPromptStore rotates through seed prompts daily; used in tests and dev.
type MemoryPromptStore struct {
	prompts []Prompt
}

func NewMemoryPromptStore() *MemoryPromptStore {
	return &MemoryPromptStore{prompts: seedPrompts}
}

func (m *MemoryPromptStore) GetTodayPrompt() (*Prompt, error) {
	if len(m.prompts) == 0 {
		return nil, nil
	}
	idx := int(time.Now().UTC().Unix()/86400) % len(m.prompts)
	p := m.prompts[idx]
	p.Date = time.Now().UTC().Truncate(24 * time.Hour)
	return &p, nil
}
