// Package kronk provides Kronk SDK lifecycle management.
// It is a general-purpose wrapper around the ardanlabs/kronk SDK and is not
// specific to any application.
//
// It handles library installation, model downloading, and engine
// initialization through a single Load call. Global SDK initialization
// is guarded by sync.Once so it is safe to call Load multiple times.
package kronk
