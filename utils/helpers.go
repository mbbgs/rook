package utils 


import (
    "os"
    "crypto/rand"
    "fmt"
    "path/filepath"
    "strings"
    "unicode"
    "math/big"
)


func GetSessionDir() (string, error) {
    cacheBase, err := os.UserCacheDir()
    if err != nil {
        return "", err
    }
    ROOK_DIR := ".rook"
    sessionDir := filepath.Join(cacheBase,ROOK_DIR)

    err = os.MkdirAll(sessionDir, 0o700)
    if err != nil {
        return "", err
    }

    // 0700 permission
    err = os.Chmod(sessionDir, 0o700)
    if err != nil {
        return "", err
    }

    return sessionDir, nil
}


func FileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}



func ValidatePassword(pw string) error {
    if len(pw) < 12 || len(pw) > 128 {
        return fmt.Errorf("password must be between 12 and 128 characters")
    }

    var hasUpper, hasLower, hasDigit, hasSymbol bool
    for _, r := range pw {
        switch {
        case unicode.IsUpper(r):
            hasUpper = true
        case unicode.IsLower(r):
            hasLower = true
        case unicode.IsDigit(r):
            hasDigit = true
        case unicode.IsPunct(r) || unicode.IsSymbol(r):
            hasSymbol = true
        }
    }

    if !(hasUpper && hasLower && hasDigit && hasSymbol) {
        return fmt.Errorf("password must include upper, lower, digit, and special character")
    }

    if strings.ContainsAny(pw, "O0l1I") {
        return fmt.Errorf("password contains ambiguous characters: O, 0, l, 1, or I")
    }

    return nil
}




func Generate(length int) (string, error) {
	srcString := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef0123456789"
	srcLen := int64(len(srcString))

	var result strings.Builder
	for i := 0; i < length; i++ {
		// Generate a random number within the length of srcString
		index, err := rand.Int(rand.Reader, big.NewInt(srcLen))
		if err != nil {
			return "", err
		}
		// Use index as the position in srcString
		result.WriteByte(srcString[index.Int64()])
	}
	return result.String(), nil
}