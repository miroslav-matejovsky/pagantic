package prompt

// ContextPolicy controls what context is included when building prompts.
type ContextPolicy struct {
	MaxTokens     int
	IncludeTools  bool
	IncludeSchema bool
}
