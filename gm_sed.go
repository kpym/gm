package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rwtodd/Go.Sed/sed"
)

// sedEngines holds the sed engines for markdown and HTML.
var (
	sedMdEngine   *sed.Engine
	sedHtmlEngine *sed.Engine
)

// addSedCommands creates or updates a sed.Engine with the given commands or file.
func addSedCommands(engine **sed.Engine, commands []string) error {
	var combinedCommands strings.Builder

	for _, cmd := range commands {
		if _, err := os.Stat(cmd); err == nil {
			// If the command is a file, load it as a sed script.
			file, err := os.Open(cmd)
			if err != nil {
				return fmt.Errorf("failed to open sed file %s: %w", cmd, err)
			}
			defer file.Close()
			if _, err := io.Copy(&combinedCommands, file); err != nil {
				return fmt.Errorf("failed to read sed file %s: %w", cmd, err)
			}
		} else {
			// Otherwise, treat it as an inline sed command.
			combinedCommands.WriteString(cmd + "\n")
		}
	}

	if combinedCommands.Len() > 0 {
		var err error
		*engine, err = sed.New(strings.NewReader(combinedCommands.String()))
		if err != nil {
			return fmt.Errorf("failed to create sed engine: %w", err)
		}
	}

	return nil
}
