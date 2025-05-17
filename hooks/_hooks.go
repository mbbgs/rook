package hooks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mbbgs/rook/consts"
	"github.com/mbbgs/rook/events"
	"github.com/mbbgs/rook/terms"
	"github.com/mbbgs/rook/utils"
	"github.com/mbbgs/rook/views"
	
	
	"golang.org/x/crypto/ssh/terminal"
)

// Event is the global event emitter instance
var Event = events.Emitter()

// SetupEventHandlers registers all event handlers when the application starts
func SetupEventHandlers() {
	// APP_BOOT event handler
	Event.On(consts.APP_BOOT, func(_ ...interface{}) {
		utils.SilentDone(consts.APP_BOOT)
		Event.Emit(consts.APP_INIT, nil)
	})

	// APP_INIT event handler
	Event.On(consts.APP_INIT, func(_ ...interface{}) {
		utils.SilentDone(consts.APP_INIT)
		Event.Off(consts.APP_BOOT)

		terms.DisplayAgreement() // Show Agreement to user
		Event.Emit(consts.APP_READY)
	})

	// APP_READY event handler
	Event.On(consts.APP_READY, func(_ ...interface{}) {
		utils.SilentDone(consts.APP_READY)
		Event.Off(consts.APP_BOOT)

		exists := checkSessionFile()

		if !exists {
			Event.Emit(consts.USER_REGISTRATION, nil)
			return
		}
		Event.Emit(consts.USER_LOGIN, nil)
	})

	// USER_LOGIN event handler
	Event.On(consts.USER_LOGIN, func(_ ...interface{}) {
		fmt.Println("[ Login Cell ]\n")
		// Use bufio to securely handle the password input
		username := promptForInput("Enter your username: ")
		password := promptForPassword("Enter your password: ")

		UserLogin(username, password, Event)
	})

	// USER_REGISTRATION event handler
	Event.On(consts.USER_REGISTRATION, func(_ ...interface{}) {
		fmt.Println("[ Registration Cell ]\n")
		// Use bufio to securely handle the password input
		username := promptForInput("Enter your username: ")
		password := promptForPassword("Enter your password: ")

		UserRegistration(username, password, Event)
	})

	// RESET_PASSWORD event handler
	Event.On(consts.RESET_PASSWORD, func(_ ...interface{}) {
		// Use bufio to securely handle the current and new password input
		currPassword := promptForPassword("Enter your current password: ")
		newPassword := promptForPassword("Enter your new password: ")

		ResetPassword(currPassword, newPassword, Event)
	})

	// DROP_TABLE event handler
	Event.On(consts.DROP_TABLE, func(_ ...interface{}) {
		// Use bufio to securely handle the password input
		currPassword := promptForPassword("Enter your password to confirm action: ")
		DropStorage(currPassword)
	})

	/*
		Event.On(consts.USER_LOGOUT, func(_ ...interface{}) {
			// Logout procedure
			session := &types.Session{}  // Get session data
			UserLogout(session)
		})
	*/

	// USER_LOGGED_IN event handler
	Event.On(consts.USER_LOGGED_IN, func(_ ...interface{}) {
		utils.SilentDone(consts.USER_LOGIN)
		Event.Off(consts.USER_LOGIN)

	dash, err := views.NewDashboard()
	if err != nil {
	utils.Error("Failed to init dashboard: " + err.Error())
	return
	}
	dash.Start()
	})
}

// Helper function to read input securely
func promptForInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// Helper function to read password securely
func promptForPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		utils.ErrorE(err)
		return ""
	}
	fmt.Println() // To move to a new line after password input
	return string(bytePassword)
}

// checkSessionFile checks if the session file exists
func checkSessionFile() bool {
	dir, err := utils.GetSessionDir()
	if err != nil {
		utils.ErrorE(err)
		return false
	}
	filePath := filepath.Join(dir, consts.STORE_FILE_PATH)
	return utils.FileExists(filePath)
}
