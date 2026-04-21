package ports

import (
	"context"
	"sync"
)

// Stage represents a single processing step in a scan pipeline.
type Stage func(ctx context.Context, in <-chan PortEntry) <-chan PortEntry

// Pipeline chains multiple Stage functions together, passing the output
// of each stage as the input to the next.
type Pipeline struct {
	stages []Stage
	mu     sync.Mutex
}

// NewPipeline returns an empty Pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// AddStage appends a processing stage to the pipeline.
func (p *Pipeline) AddStage(s Stage) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stages = append(p.stages, s)
}

// Run feeds entries into the first stage and returns the final output
// channel. If no stages have been added the source channel is returned
// unchanged.
func (p *Pipeline) Run(ctx context.Context, source <-chan PortEntry) <-chan PortEntry {
	p.mu.Lock()
	stages := make([]Stage, len(p.stages))
	copy(stages, p.stages)
	p.mu.Unlock()

	current := source
	for _, stage := range stages {
		current = stage(ctx, current)
	}
	return current
}

// FilterStage returns a Stage that drops entries for which keep returns false.
func FilterStage(keep func(PortEntry) bool) Stage {
	return func(ctx context.Context, in <-chan PortEntry) <-chan PortEntry {
		out := make(chan PortEntry)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-in:
					if !ok {
						return
					}
					if keep(e) {
						select {
						case out <- e:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}()
		return out
	}
}

// TransformStage returns a Stage that applies fn to every entry.
func TransformStage(fn func(PortEntry) PortEntry) Stage {
	return func(ctx context.Context, in <-chan PortEntry) <-chan PortEntry {
		out := make(chan PortEntry)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case e, ok := <-in:
					if !ok {
						return
					}
					select {
					case out <- fn(e):
					case <-ctx.Done():
						return
					}
				}
			}
		}()
		return out
	}
}
