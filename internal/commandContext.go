package internal

// defines the interface for all strategies
type CommandStrategy interface {
	Execute(args []string) error
}

type CommandContext struct {
	strategy CommandStrategy
}

func (c *CommandContext) SetStrategy(strategy CommandStrategy) {
	c.strategy = strategy
}

func (c *CommandContext) Execute(args []string) error {
	return c.strategy.Execute(args)
}
