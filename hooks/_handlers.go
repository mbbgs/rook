package hooks

import (
	"fmt"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/mbbgs/rook-go/consts"
	"github.com/mbbgs/rook-go/events"
	"github.com/mbbgs/rook-go/securecrypto"
	"github.com/mbbgs/rook-go/terms"
	"github.com/mbbgs/rook-go/types"
	"github.com/mbbgs/rook-go/utils"
	"github.com/mbbgs/rook-go/store"
)

func UserRegistration(username, password, masterkey string, Event *events.Event) {
	username, password , masterkey = sanitizeCreds(username, password,masterkey)

	if !isValidCreds(username, password, masterkey) || !validatePassword(password) {
		return
	}
	msalt, err := securecrypto.GenerateSalt(consts.SALT_SIZE)
	if err != nil {
		utils.ErrorE(err)
		return
	}
	
	salt, err := securecrypto.GenerateSalt(consts.SALT_SIZE)
	if err != nil {
		utils.ErrorE(err)
		return
	}

	_, hashedPassword, err := securecrypto.HashWithSalt(password, salt)
	if err != nil {
		utils.ErrorE(err)
		return
	}
	_,hashedMaster, err := securecrypto.HashWithSalt(masterkey,msalt)
	if err != nil {
		utils.ErrorE(err)
		return
	}
	passSalt := fmt.Sprintf("%s:%s",hashedPassword,salt)
	masterSalt := fmt.Sprintf("%s:%s",hashedMaster,msalt)

	/*	_ = utils.SaveFileToSessionDir(consts.SECRET_ROOK, salt)

	newUser := types.NewUser(username, hashedPassword)
	session := types.NewSession(newUser)
	session.SaveSession()
	*/
	newUser := models.NewUser(username,passSalt,masterSalt)
	store, err := store.NewStore()
	
	if err != nil {
		utils.ErrorE(err)
		return
	}
	
	defer store.Close()
	
	store.CreaterUser(newUser)
	store.Save()
	
	Event.Emit(consts.USER_LOGIN, nil)
}












func UserLogin(username, password string, Event *events.Event) {
	username, password = sanitizeCreds(username, password,nil)
	if !isValidCreds(username, password) || !validatePassword(password) {
		return
	}

	attemptFile := getAttemptsFilePath()
	attempts := readAttempts(attemptFile)
	if handleExcessiveAttempts(attempts, attemptFile) {
		return
	}

	session, salt, ok := loadSessionAndSalt(attemptFile)
	if !ok {
		return
	}

	if session.User.Username != username {
		failAttempt("Invalid username or password.", attemptFile, attempts)
		return
	}

	_, hashedPassword, err := securecrypto.HashWithSalt(password, salt)
	if err != nil {
		utils.ErrorE(err)
		return
	}
	if !bytes.Equal([]byte(session.User.Password), []byte(hashedPassword)) {
		failAttempt("Invalid username or password.", attemptFile, attempts)
		return
	}

	os.Remove(attemptFile)

	newSalt, _ := securecrypto.GenerateSalt(consts.SALT_SIZE)
	_ = utils.SaveFileToSessionDir(consts.SECRET_ROOK, newSalt)

	_, newHash, _ := securecrypto.HashWithSalt(password, newSalt)
	session.User.Password = newHash
	session.UpdateLastAccess()
	session.SaveSession()

	Event.Emit(consts.USER_LOGGED_IN, nil)
}

func ResetPassword(oldPassword, newPassword string, Event *events.Event) {
	oldPassword, newPassword = sanitizeCreds(oldPassword, newPassword)
	if !isValidCreds(oldPassword, newPassword) || !validatePassword(oldPassword) || !validatePassword(newPassword) {
		return
	}

	attemptFile := getAttemptsFilePath()
	attempts := readAttempts(attemptFile)
	if handleExcessiveAttempts(attempts, attemptFile) {
		utils.Warn("Login now.")
		return
	}

	session, salt, ok := loadSessionAndSalt(attemptFile)
	if !ok {
		return
	}

	_, hashedOld, _ := securecrypto.HashWithSalt(oldPassword, salt)
	if !bytes.Equal([]byte(session.User.Password), []byte(hashedOld)) {
		failAttempt("Invalid credentials provided.", attemptFile, attempts)
		return
	}

	os.Remove(attemptFile)

	newSalt, _ := securecrypto.GenerateSalt(consts.SALT_SIZE)
	_ = utils.SaveFileToSessionDir(consts.SECRET_ROOK, newSalt)

	_, newHash, _ := securecrypto.HashWithSalt(newPassword, newSalt)
	session.User.Password = newHash
	session.UpdateLastAccess()
	session.SaveSession()

	Event.Emit(consts.USER_LOGIN, nil)
}

func DropStorage(currentPassword string) {
	currentPassword = strings.TrimSpace(currentPassword)
	if currentPassword == "" {
		utils.Warn("Please provide the current password.")
		os.Exit(1)
	}

	attemptFile := getAttemptsFilePath()
	attempts := readAttempts(attemptFile)
	if handleExcessiveAttempts(attempts, attemptFile) {
		return
	}

	session, salt, ok := loadSessionAndSalt(attemptFile)
	if !ok {
		os.Exit(1)
	}

	_, hashedInput, _ := securecrypto.HashWithSalt(currentPassword, salt)
	if !bytes.Equal([]byte(session.User.Password), []byte(hashedInput)) {
		failAttempt("Invalid current password.", attemptFile, attempts)
		os.Exit(1)
	}

	terms.NukeFiles()
	os.Remove(attemptFile)

	utils.Done("Session nuked and files deleted successfully.")
	os.Exit(1)
}

func UserLogout(session *types.Session) {
	session.UpdateLastAccess()
	session.SaveSession()
	os.Exit(1)
}

// ---------- Helpers ----------

func sanitizeCreds(u, p , m string) (string, string) {
	return strings.TrimSpace(u), strings.TrimSpace(p),strings.TrimSpace(m)
}

func isValidCreds(u, p , m string) bool {
	if u == "" || p == "" || m == "" {
		utils.Warn("Provide valid credentials.")
		return false
	}
	if len(p) < 8 {
		utils.Warn("Password must be at least 8 characters long.")
		return false
	}
		if len(m) < 8 {
		utils.Warn("Master key must be at least 8 characters long.")
		return false
	}
	return true
}

func validatePassword(p string) bool {
	if err := utils.ValidatePassword(p); err != nil {
		utils.ErrorE(err)
		return false
	}
	return true
}

func getAttemptsFilePath() string {
	dir, err := utils.GetSessionDir()
	if err != nil {
		utils.ErrorE(err)
		return ""
	}
	return filepath.Join(dir, consts.ATTEMPTS_PATH)
}

func readAttempts(path string) int {
	var attempts int
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &attempts)
	}
	return attempts
}

func saveAttempts(path string, count int) {
	data, _ := json.Marshal(count)
	_ = os.WriteFile(path, data, 0600)
}

func failAttempt(msg, path string, count int) {
	utils.Warn(msg)
	count++
	saveAttempts(path, count)
}

func handleExcessiveAttempts(attempts int, path string) bool {
	if attempts == consts.MAX_M_ATTEMPTS {
		utils.Warn("Too many failed attempts. Further tries would lead to Zero Trust.")
		return true
	}
	if attempts >= int(consts.MAX_ATTEMPTS+1) {
		terms.NukeFiles()
		return true
	}
	return false
}


func loadSessionAndSalt(attemptFile string) (*types.Session, []byte, bool) {
	sbytes, err := utils.ReadFileFromSessionDir(consts.AUTH_FILE_PATH)
	if err != nil {
		utils.ErrorE(err)
		failAttempt("Could not read session.", attemptFile, readAttempts(attemptFile))
		return nil, nil, false
	}

	secret, err := utils.ReadFileFromSessionDir(consts.SECRET_ROOK)
	if err != nil {
		utils.ErrorE(err)
		failAttempt("Could not read secret.", attemptFile, readAttempts(attemptFile))
		return nil, nil, false
	}

	var session types.Session
	if err := json.Unmarshal(sbytes, &session); err != nil {
		utils.ErrorE(err)
		failAttempt("Corrupted session file.", attemptFile, readAttempts(attemptFile))
		return nil, nil, false
	}
	return &session, secret, true
}