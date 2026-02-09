package executor

import (
	"context"
)

// Command is the unit of work in a pipeline
type Command interface {
	Run(ctx context.Context) error
}

type command struct {
	run       func(ctx context.Context) error
	preHooks  []func(ctx context.Context) error
	postHooks []func(ctx context.Context) error
}

func (c *command) Run(ctx context.Context) error {
	for _, h := range c.preHooks {
		if err := h(ctx); err != nil {
			return err
		}
	}
	if err := c.run(ctx); err != nil {
		return err
	}
	for _, h := range c.postHooks {
		if err := h(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c *command) AddPreHook(h func(ctx context.Context) error) *command {
	c.preHooks = append(c.preHooks, h)
	return c
}

func (c *command) AddPostHook(h func(ctx context.Context) error) *command {
	c.postHooks = append(c.postHooks, h)
	return c
}

// CommandFunc is now a factory
func CommandFunc(fn func(context.Context) error) *command {
	return &command{run: fn}
}
