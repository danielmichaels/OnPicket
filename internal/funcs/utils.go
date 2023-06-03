package funcs

import (
	"github.com/rs/zerolog/log"
	"os"
	"runtime/debug"
	"sync"
)

// GetEnv improves on os.Getenv by setting a default value if the variable
// is not set
func GetEnv(val, defaultVal string) string {
	if os.Getenv(val) == "" {
		return defaultVal
	}
	return os.Getenv(val)
}

// BackgroundFunc executes a given function in a goroutine.
func BackgroundFunc(fn func()) {
	// Launch a background goroutine.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		// Recover any panic.
		defer func() {
			if err := recover(); err != nil {
				log.Warn().Msgf("background-task err: %v", err)
				debug.PrintStack()
			}
		}()

		// Execute the arbitrary function that we passed as the parameter.
		wg.Done()
		fn()
	}()
	wg.Wait()
}
