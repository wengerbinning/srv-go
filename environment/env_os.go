package environment

import (
	"os"
	"fmt"
)

func ensureFold(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	return nil
}
