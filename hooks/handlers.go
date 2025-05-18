package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mbbgs/rook/consts"
	"github.com/mbbgs/rook/events"
	"github.com/mbbgs/rook/models"
	"github.com/mbbgs/rook/securecrypto"
	"github.com/mbbgs/rook/store"
	"github.com/mbbgs/rook/terms"
	"github.com/mbbgs/rook/utils"
)

func UserRegistration(username, password, masterkey string, Event *events.Event) {
	username, password, masterkey = sanitizeCreds(username, password, masterkey)
	if !isValidCreds(username, password, masterkey) || !validatePassword(password) {
		return
	}

	salt, err := securecrypto.GenerateSalt()
	if err != nil {
		utils.ErrorE(err)
		return
	}

	msalt, err := securecrypto.GenerateSalt()
	if err != nil {
		utils.ErrorE(err)
		return
	}

	 hashedPassword, err := securecrypto.HashWithSalt(password, salt)
	if err != nil {
		utils.ErrorE(err)
		return
	}

	hashedMaster, err := securecrypto.HashWithSalt(masterkey, msalt)
	if err != nil {
		utils.ErrorE(err)
		return
	}

	newUser := models.NewUser(
		username,
		fmt.Sprintf("%s:%s", hashedPassword, salt),
		fmt.Sprintf("%s:%s", hashedMaster, msalt),
	)

	db, err := store.NewStore()
	if err != nil {
		utils.ErrorE(err)
		return
	}
	defer db.Close()

	if err := db.CreaterUser(newUser); err != nil {
		utils.ErrorE(err)
		return
	}

	
	Event.Emit(consts.USER_LOGIN, nil)
}

func UserLogin(username, password string, Event *events.Event) {
	username, password, _ = sanitizeCreds(username, password, "")
	if !isValidCreds(username, password, "") || !validatePassword(password) {
		return
	}

	attemptPath := getAttemptsFilePath()
	attempts := readAttempts(attemptPath)
	if handleExcessiveAttempts(attempts, attemptPath) {
		return
	}

	db, err := store.NewStore()
	if err != nil {
		utils.ErrorE(err)
		failAttempt("Failed to open database.", attemptPath, attempts)
		return
	}
	user, err := db.GetUser(username)
	if err != nil {
		failAttempt("Invalid username or password.", attemptPath, attempts)
		return
	}

	storedHash, salt, ok := splitHashSalt(user.Password)
	if !ok {
		failAttempt("Invalid password format.", attemptPath, attempts)
		return
	}

	 inputHash, err := securecrypto.HashWithSalt(password,salt)
	if err != nil || !bytes.Equal([]byte(storedHash), []byte(inputHash)) {
		failAttempt("Invalid username or password.", attemptPath, attempts)
		return
	}

	_ = os.Remove(attemptPath)
	utils.Done("User logged in successfully.")
	Event.Emit(consts.USER_LOGGED_IN, &user, &db)
}

func ResetPassword(username, oldPassword, newPassword string, Event *events.Event) {
	username, oldPassword, newPassword = sanitizeCreds(username, oldPassword, newPassword)
	if !isValidCreds(username, oldPassword, "") || !validatePassword(oldPassword) || !validatePassword(newPassword) {
		return
	}

	attemptPath := getAttemptsFilePath()
	attempts := readAttempts(attemptPath)
	if handleExcessiveAttempts(attempts, attemptPath) {
		return
	}

	db, err := store.NewStore()
	if err != nil {
		utils.ErrorE(err)
		return
	}
	defer db.Close()

	user, err := db.GetUser(username)
	if err != nil {
		failAttempt("Invalid credentials.", attemptPath, attempts)
		return
	}

	oldHash, oldSalt, ok := splitHashSalt(user.Password)
	if !ok {
		failAttempt("Invalid stored password.", attemptPath, attempts)
		return
	}

	checkOld, err := securecrypto.HashWithSalt(oldPassword,oldSalt)
	if err != nil || !bytes.Equal([]byte(oldHash), []byte(checkOld)) {
		failAttempt("Old password incorrect.", attemptPath, attempts)
		return
	}

	newSalt, err := securecrypto.GenerateSalt()
	if err != nil {
		utils.ErrorE(err)
		return
	}

	newHash, err := securecrypto.HashWithSalt(newPassword, newSalt)
	if err != nil {
		utils.ErrorE(err)
		return
	}

	user.Password = []byte(fmt.Sprintf("%s:%s", newHash, newSalt))
	if err := db.UpdateUser(user); err != nil {
		utils.ErrorE(err)
		return
	}



	_ = os.Remove(attemptPath)
	utils.Done("Password reset successful.")
	Event.Emit(consts.USER_LOGIN, nil)
}

func DropStorage(username, currentPassword string) {
	username = strings.TrimSpace(username)
	currentPassword = strings.TrimSpace(currentPassword)
	if username == "" || currentPassword == "" {
		utils.Warn("Username and password required.")
		return
	}

	attemptPath := getAttemptsFilePath()
	attempts := readAttempts(attemptPath)
	if handleExcessiveAttempts(attempts, attemptPath) {
		return
	}

	db, err := store.NewStore()
	if err != nil {
		utils.ErrorE(err)
		return
	}
	defer db.Close()

	user, err := db.GetUser(username)
	if err != nil {
		failAttempt("Authentication failed.", attemptPath, attempts)
		return
	}

	storedHash, salt, ok := splitHashSalt(user.Password)
	if !ok {
		failAttempt("Stored password invalid.", attemptPath, attempts)
		return
	}

	_, checkHash, err := securecrypto.HashWithSalt(currentPassword, []byte(salt))
	if err != nil || !bytes.Equal([]byte(storedHash), []byte(checkHash)) {
		failAttempt("Invalid password.", attemptPath, attempts)
		return
	}

	terms.NukeFiles()
	_ = os.Remove(attemptPath)
	utils.Done("All storage nuked successfully.")
}

func UserLogout() {
	utils.Done("User logged out.")
}

// --------- Helpers ---------

func sanitizeCreds(u, p, m string) (string, string, string) {
	return strings.TrimSpace(u), strings.TrimSpace(p), strings.TrimSpace(m)
}

func isValidCreds(u, p, m string) bool {
	if u == "" || p == "" {
		utils.Warn("Provide username and password.")
		return false
	}
	if len(p) < 8 {
		utils.Warn("Password must be at least 8 characters.")
		return false
	}
	if m != "" && len(m) < 8 {
		utils.Warn("Masterkey must be at least 8 characters.")
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
		utils.Warn("Too many failed attempts.")
		return true
	}
	if attempts >= int(consts.MAX_ATTEMPTS+1) {
		terms.NukeFiles()
		return true
	}
	return false
}

func splitHashSalt(combined []byte) (hash string, salt string, ok bool) {
	parts := strings.Split(string(combined), ":")
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}