package terms

import (
  "fmt"
  "os"
  "math/rand"
  
  "github.com/mbbgs/rook-go/consts"
)

func NukeFiles() {
    files := []string{consts.SECRET_ROOK, consts.AUTH_FILE_PATH, consts.STORE_FILE_PATH, consts.ATTEMPTS_PATH}

    for _, file := range files {
        if _, err := os.Stat(file); os.IsNotExist(err) {
            continue
        }

        if err := ShredAndDelete(file, 3); err != nil {
            fmt.Printf("Failed to shred %s: %v\n", file, err)
        }
    }

    fmt.Println("All sensitive files shredded and wiped.")
    os.Exit(1)
}

func ShredAndDelete(path string, passes int) error {
    info, err := os.Stat(path)
    if err != nil {
        return err
    }

    size := info.Size()
    file, err := os.OpenFile(path, os.O_WRONLY, 0600)
    if err != nil {
        return err
    }
    defer file.Close()

    random := make([]byte, size)
    for i := 0; i < passes; i++ {
        if _, err := rand.Read(random); err != nil {
            return err
        }
        if _, err := file.WriteAt(random, 0); err != nil {
            return err
        }
    }

    file.Sync()
    file.Close()

    return os.Remove(path)
}