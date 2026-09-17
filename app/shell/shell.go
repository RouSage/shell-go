package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type Shell struct {
	history     *History
	jobs        *Jobs
	completions map[string]string
}

func New() *Shell {
	return &Shell{
		history:     NewHistory(),
		jobs:        NewJobs(),
		completions: map[string]string{},
	}
}

func (s *Shell) REPL() {
	completer := NewCompleter(s)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "$ ",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating readline: %v\n", err)
		return
	}
	defer func() {
		if err := rl.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing readline: %v\n", err)
		}
	}()

	for {
		// Read the command
		command, err := rl.Readline()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}

		pipeline := NewPipeline(s, command)
		if pipeline == nil {
			continue
		}

		pipeline.Run()
	}
}
