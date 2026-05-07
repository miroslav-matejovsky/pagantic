package kronk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ardanlabs/kronk/sdk/kronk"
	"github.com/ardanlabs/kronk/sdk/kronk/model"
	"github.com/ardanlabs/kronk/sdk/tools/defaults"
	"github.com/ardanlabs/kronk/sdk/tools/libs"
	"github.com/ardanlabs/kronk/sdk/tools/models"
)

// installTimeout limits how long library and model downloads can take.
const installTimeout = 25 * time.Minute

// Config controls engine initialization.
type Config struct {
	// ModelSource is a HuggingFace URL, canonical "provider/modelID",
	// or bare model id. Required.
	ModelSource string
}

var (
	initOnce sync.Once
	initErr  error // set once by initOnce; non-nil means Init failed permanently.
)

// Load downloads Kronk runtime libraries and model files, initializes
// the inference engine, and returns the ready-to-use Kronk instance.
// The returned cleanup function unloads the model and should be deferred.
func Load(ctx context.Context, cfg Config) (*kronk.Kronk, func(), error) {
	if cfg.ModelSource == "" {
		return nil, nil, fmt.Errorf("kronk: Config.ModelSource must not be empty")
	}

	mp, err := install(ctx, cfg.ModelSource)
	if err != nil {
		return nil, nil, fmt.Errorf("kronk install: %w", err)
	}

	krn, err := newEngine(mp)
	if err != nil {
		return nil, nil, fmt.Errorf("kronk engine: %w", err)
	}

	cleanup := func() {
		fmt.Println("\nUnloading model")
		if err := krn.Unload(context.Background()); err != nil {
			fmt.Printf("Failed to unload model: %v\n", err)
		}
	}

	return krn, cleanup, nil
}

// install downloads Kronk runtime libraries and LLM model files.
func install(ctx context.Context, modelSource string) (models.Path, error) {
	fmt.Println("Installing kronk system (libs and models)")

	ctx, cancel := context.WithTimeout(ctx, installTimeout)
	defer cancel()

	l, err := libs.New(
		libs.WithVersion(defaults.LibVersion("")),
	)
	if err != nil {
		return models.Path{}, fmt.Errorf("create libs system: %w", err)
	}

	if version, err := l.Download(ctx, kronk.FmtLogger); err != nil {
		return models.Path{}, fmt.Errorf("install libs: %w", err)
	} else {
		fmt.Printf("Installed libs version: %+v\n", version)
	}

	mdls, err := models.New()
	if err != nil {
		return models.Path{}, fmt.Errorf("create models system: %w", err)
	}

	fmt.Println("Downloading model:", modelSource)

	mp, err := mdls.Download(ctx, kronk.FmtLogger, modelSource)
	if err != nil {
		return models.Path{}, fmt.Errorf("install model: %w", err)
	}

	fmt.Printf("Model downloaded to: %s\n", mp.ModelFiles[0])
	return mp, nil
}

// newEngine initializes the Kronk inference engine with downloaded model files.
func newEngine(mp models.Path) (*kronk.Kronk, error) {
	fmt.Println("Loading model...")

	initOnce.Do(func() {
		initErr = kronk.Init()
	})
	if initErr != nil {
		return nil, fmt.Errorf("init kronk: %w", initErr)
	}

	krn, err := kronk.New(
		model.WithModelFiles(mp.ModelFiles),
	)
	if err != nil {
		return nil, fmt.Errorf("create inference model: %w", err)
	}

	fmt.Println("  Model type     :", krn.ModelInfo().Type)
	fmt.Println("  Context window :", krn.ModelConfig().ContextWindow())
	fmt.Println("  Flash attention:", krn.ModelConfig().FlashAttention)
	fmt.Println("  Model size     :", krn.ModelInfo().Size/(1000*1000), "MB")
	fmt.Println()

	return krn, nil
}
