package hooks

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mbbgs/rook/consts"
	"github.com/mbbgs/rook/events"
	"github.com/mbbgs/rook/terms"
	"github.com/mbbgs/rook/utils"
	"github.com/mbbgs/rook/views"
	"github.com/mbbgs/rook/store"
	
	"golang.org/x/crypto/ssh/terminal"
)

// Event is the global event emitter instance
var Event = events.Emitter()

// SetupEventHandlers registers all event handlers when the application starts
func SetupEventHandlers() {
	Event.On(consts.APP_BOOT, func(_ ...interface{}) {
		utils.SilentDone(consts.APP_BOOT)
		Event.Emit(consts.APP_INIT,nil)
	})

	Event.On(consts.APP_INIT, func(_ ...interface{}) {
		utils.SilentDone(consts.APP_INIT)
		Event.Off(consts.APP_BOOT)

		terms.DisplayAgreement()
		Event.Emit(consts.APP_READY,nil)
	})
/*
	Event.On(consts.APP_READY, func(_ ...interface{}) {
		utils.SilentDone(consts.APP_READY)
		Event.Off(consts.APP_INIT)

		exists, err := store.Exists()
		if err != nil {
			utils.Error("Failed to check user existence: " + err.Error())
			os.Exit(1)
		}

		if !exists {
			Event.Emit(consts.USER_REGISTRATION)
			return
		}

		Event.Emit(consts.USER_LOGIN)
	})
	*/
	
	Event.On(consts.APP_READY, func(_ ...interface{}) {
	utils.SilentDone(consts.APP_READY)
	Event.Off(consts.APP_INIT)
	loadStore := store.NewStore()
	exists, err := loadStore.IsUser()
	if err != nil {
		utils.Error("Failed to check user existence: " + err.Error())
		os.Exit(1)
	}

	if !exists {
		Event.Emit(consts.USER_REGISTRATION,nil)
		return
	}

	Event.Emit(consts.USER_LOGIN,nil)
})

	Event.On(consts.USER_LOGIN, func(_ ...interface{}) {
		fmt.Println("[ Login Cell ]\n")

	//	for attempts := 0; attempts < 3; attempts++ {
			username, password := promptForCredentials()
			 UserLogin(username, password, Event) 
			//fmt.Println("Invalid credentials. Try again.")
	//	}

	//	fmt.Println("Too many failed attempts.")
		//os.Exit(1)
	})

	Event.On(consts.USER_REGISTRATION, func(_ ...interface{}) {
		fmt.Println("[ Registration Cell ]\n")
		username, password := promptForCredentials()
		masterKey := promptForInput("choose your Master key: ")

		UserRegistration(username, password,masterKey, Event)
	})

	Event.On(consts.RESET_PASSWORD, func(_ ...interface{}) {
		username := promptForPassword("Enter your username: ")
		currPassword := promptForPassword("Enter your current password: ")
	
		newPassword := promptForPassword("Enter your new password: ")
		ResetPassword(username,currPassword, newPassword, Event)
	})

	Event.On(consts.DROP_TABLE, func(_ ...interface{}) {
		username := strings.TrimSpace(Event.Username)
		if username == "" {
			Event.Emit(consts.F_USER_LOGOUT,nil)
			return
		}
		currPassword := promptForPassword("Enter your password to confirm action: ")
		DropStorage(username,currPassword)
	})

	Event.On(consts.USER_LOGGED_IN, func(args ...interface{}) {
		utils.SilentDone(consts.USER_LOGIN)
		Event.Off(consts.USER_LOGIN)
		
		dash, err := dashboard.NewDashboard(args[0],args[1])
		//	if err != nil {
		//	utils.Error("Failed to init dashboard: " + err.Error())
		//	return
		//	}
		dash.Start()
	})
	
	Event.On(consts.F_USER_LOGOUT, func(_ ...interface{}) {
		UserLogout()
		fmt.Println("Session expired")
		Event.Emit(consts.USER_LOGIN,nil)
	})
	
	
	Event.On(consts.USER_LOGOUT, func(_ ...interface{}) {
		UserLogout()
		fmt.Println("Logged out successfully.")
		Event.Emit(consts.USER_LOGIN,nil)
	})
	
}

// promptForCredentials prompts for both username and password
func promptForCredentials() (string, string) {
	username := promptForInput("Enter your username: ")
	password := promptForPassword("Enter your password: ")
	return username, password
}

// promptForInput reads input from stdin with prompt
func promptForInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptForPassword reads password from stdin securely
func promptForPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		utils.ErrorE(err)
		return ""
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword))
}

