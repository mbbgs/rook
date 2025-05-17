package main

import (
	"flag"
	"os"
	"fmt"
	"github.com/mbbgs/rook/utils"
	"github.com/mbbgs/rook/consts"
	"github.com/mbbgs/rook/hooks"
)

func main() {
	// Setup event handlers before using them
	hooks.SetupEventHandlers()
	if err := utils.InitLogger(); err != nil {
		fmt.Println()
	}
	defer utils.CloseLogger()
	
	// Parse command line flags
	reset := flag.Bool("reset", false, "Reset your password")
	drop := flag.Bool("drop", false, "Permanently delete your encrypted storage")
	flag.Parse()

	// Handle command line options
switch {
	case *reset:
		hooks.Event.Emit(consts.RESET_PASSWORD, nil)
		waitForExit()
		return
	case *drop:
		hooks.Event.Emit(consts.DROP_TABLE, nil)
		waitForExit()
		return
	default:
		hooks.Event.Emit(consts.APP_BOOT, nil)
	}
}

func waitForExit() {
	println("\nPress CTRL + SPACE to exit.")
	b := make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if b[0] == 0 { // CTRL + SPACE = ASCII NULL
			break
		}
	}
}
