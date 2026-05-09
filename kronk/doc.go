// Package kronk provides Kronk SDK lifecycle management and inference adapter.
//
// It is a general-purpose wrapper around the ardanlabs/kronk SDK and is not
// specific to any application.
//
// It handles library installation, model downloading, and engine
// initialization through a single Load call. Config.ModelSource is required
// and must specify the model to load. Global SDK initialization is guarded
// by sync.Once so it is safe to call Load multiple times.
//
// Adapter implements inference.Engine by converting between pagantic core
// types and kronk model.D format, keeping the inference layer free of
// vendor-specific imports.
package kronk
