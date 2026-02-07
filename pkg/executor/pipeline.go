package executor

import (
	"context"
	"log/slog"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Step is a pipeline Step.
// It could either have substeps or commands. Not both at the same time.
type Step struct {
	SubSteps []Step
	Commands []Command

	// Parallel means all the Commands/SubSteps will be executed in parallel
	Parallel bool
}

// Pipeline executes a sequence of Steps
type Pipeline struct {
	logger *slog.Logger
	mu     sync.Mutex
	cancel func()
	steps  []Step
}

func NewPipeline(logger *slog.Logger, steps []Step) *Pipeline {
	if logger == nil {
		logger = slog.Default()
	}
	return &Pipeline{logger: logger, steps: steps}
}

func (p *Pipeline) Start(parent context.Context) error {
	p.mu.Lock()
	ctx, cf := context.WithCancel(parent)
	p.cancel = cf
	defer p.mu.Unlock()

	for i := range p.steps {
		step := p.steps[i]
		if err := p.execStep(ctx, &step); err != nil {
			return err
		}
	}

	return nil
}

func (p *Pipeline) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

func (p *Pipeline) execStep(ctx context.Context, step *Step) error {
	if err := p.execSubSteps(ctx, step); err != nil {
		return err
	}
	return p.execCommands(ctx, step)
}

func (p *Pipeline) execCommands(ctx context.Context, step *Step) error {
	if step.Parallel {
		g, gctx := errgroup.WithContext(ctx)
		for i := range step.Commands {
			cmd := step.Commands[i]
			g.Go(func() error {
				return cmd.Run(gctx)
			})
		}
		return g.Wait()
	}

	for i := range step.Commands {
		if err := step.Commands[i].Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) execSubSteps(ctx context.Context, step *Step) error {
	if step.Parallel {
		g, gctx := errgroup.WithContext(ctx)
		for i := range step.SubSteps {
			substep := &step.SubSteps[i]
			g.Go(func() error {
				return p.execStep(gctx, substep)
			})
		}
		return g.Wait()
	}

	for i := range step.SubSteps {
		if err := p.execStep(ctx, &step.SubSteps[i]); err != nil {
			return err
		}
	}
	return nil
}
