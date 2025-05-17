package terms

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/term"

	"github.com/mbbgs/rook-go/consts"
)

func DisplayAgreement() {
	// Check if already agreed
	path := filepath.Join(".", consts.AGREEMENT_FILE)
	if _, err := os.Stat(path); err == nil {
		return
	}

	// Show terms
	fmt.Println("\nROOK SECURITY POLICY & TERMS")
	fmt.Println(consts.TERMS)

	// Get initial terminal state
	oldState, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Failed to get terminal state:", err)
		os.Exit(1)
	}

	// Enter raw mode
	_, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Failed to enter raw mode:", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Signal handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigs)

	go func() {
		<-sigs
		fmt.Print("\r\n")
		term.Restore(int(os.Stdin.Fd()), oldState)
		os.Exit(0)
	}()

	// Display initial prompt
	fmt.Print("TYPE \"I AGREE\" TO CONTINUE (or press Ctrl+Space to quit): ")

	var input []rune
	buf := make([]byte, 1)

	for {
		n, err := os.Stdin.Read(buf)
		if err != nil || n == 0 {
			continue
		}

		switch buf[0] {
		case 0: // Ctrl+Space
			fmt.Print("\r\n")
			term.Restore(int(os.Stdin.Fd()), oldState)
			os.Exit(0)

		case '\r', '\n': // Enter
			if strings.EqualFold(string(input), "I AGREE") {
				fmt.Print("\r\n")
				// Save agreement
				if err := os.WriteFile(path, []byte("agreed"), 0600); err != nil {
					fmt.Println("\nError signing agreement:", err)
					os.Exit(1)
				}
				return
			}
			// Clear input and reprompt
			input = input[:0]
			fmt.Print("\r\nTYPE \"I AGREE\" TO CONTINUE: ")

		case 127, 8: // Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
				fmt.Print("\b \b")
			}

		default: // Normal input
			if len(input) < 50 { // Prevent buffer overflow
				input = append(input, rune(buf[0]))
				fmt.Printf("%c", buf[0])
			}
		}
	}
}
