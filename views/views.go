package dashboard

import (
    "bufio"
    "encoding/json"
    "fmt"
    "time"
    "strings"
    "os"
    "path/filepath"
    
    "github.com/mbbgs/rook/consts"
    "github.com/mbbgs/rook/models"
    "github.com/mbbgs/rook/store"
    "github.com/mbbgs/rook/types"
    "github.com/mbbgs/rook/utils"
)

type Dashboard struct {
    storage *store.Store
    user  *models.User
}

func NewDashboard(storee any, user any) *Dashboard {
    s, ok1 := storee.(*store.Store)
    u, ok2 := user.(*models.User)
    if !ok1 || !ok2 {
        panic("Invalid types passed to NewDashboard")
    }
    return &Dashboard{storage: s, user: u}
}
func (d *Dashboard) Start() {
    scanner := bufio.NewScanner(os.Stdin)
    cmdList()
    for {
        fmt.Print("dashboard> ")
        if !scanner.Scan() {
            break
        }
        cmdLine := strings.TrimSpace(scanner.Text())
        if cmdLine == "" {
            continue
        }

        parts := strings.SplitN(cmdLine, " ", 2)
        cmd := strings.ToLower(parts[0])
        arg := ""
        if len(parts) > 1 {
            arg = strings.TrimSpace(parts[1])
        }

        switch cmd {
        case "list":
            d.listData()
        case "add":
            d.addData()
        case "get":
            if arg == "" {
                fmt.Println("Usage: get <label>")
                continue
            }
            d.getByLabel(arg)
        case "remove":
            if arg == "" {
                fmt.Println("Usage: remove <label>")
                continue
            }
            d.removeByLabel(arg)
        case "wipe":
            d.wipeStore()
        case "help":
            d.printHelp()
        case "user":
            d.getCurUser()
        case "clear":
            d.clearLog()
        case "view":
            d.viewLogs()
        case "exit", "quit":
            fmt.Println("Bye!")
            return
        default:
            fmt.Println("Unknown command:", cmd)
            d.printHelp()
        }
    }
}

func (d *Dashboard) printHelp() {
    fmt.Println(`Commands:
  list              - List all entries for current user
  add               - Add new entry
  get <label>       - Show entry by label
  remove <label>    - Remove entry by label
  wipe              - Wipe entire store (all users)
  help              - Show this help
  user              - get current user
  clear             - clear app logs
  view              - view  app logs
  exit, quit        - Exit dashboard`)
}

func (d *Dashboard) listData() {
    found := false
    allData, err := d.storage.GetAllForUser(d.user.Username)
    if err != nil {
        fmt.Println("Failed to list data:", err)
        return
    }

    for label, data := range allData {
        maskedPwd := maskPassword(string(data.Lpassword))
        fmt.Printf("[%s]\n  URL: %s\n  User: %s\n  Password: %s\n  Last Access: %s\n\n",
            label, data.Lurl, data.Lname, maskedPwd, data.LastAccess.Format(time.RFC1123))
        found = true
    }

    if !found {
        fmt.Println("No saved entries for user", d.user.Username)
    }
}

func (d *Dashboard) addData() {
    
	var label, lname, lpassword, lurl string

	fmt.Print("Enter Label: ")
	fmt.Scanln(&label)
	if label == "" {
		fmt.Println("Label cannot be empty.")
		return
	}

	fmt.Print("Enter Username/Email: ")
	fmt.Scanln(&lname)
	if lname == "" {
		fmt.Println("Username/Email cannot be empty.")
		return
	}

	fmt.Print("Enter Password: ")
	fmt.Scanln(&lpassword)
	if lpassword == "" {
		fmt.Println("Password cannot be empty.")
		return
	}

	fmt.Print("Enter URL (optional): ")
	fmt.Scanln(&lurl)
	if lurl == "" {
		lurl = "(Not Set)"
	}

	data := types.Data{
		Lname:     lname,
		Lpassword: []byte(lpassword),
		Lurl:      lurl,
	}

	err := d.storage.AddToStore(d.user.Username, types.Label(label), data)
	if err != nil {
		fmt.Println("Failed to add data:", err)
	} else {
		fmt.Println("Data added successfully.")
	}
	
    /***
    err := d.storage.AddToStore(d.user.Username, types.Label(label), data)
    if err != nil {
        fmt.Println("Failed to add data:", err)
        return
    }
    fmt.Println("Data added successfully.")
    ****/
    
}

func (d *Dashboard) getByLabel(label string) {
    data, err := d.storage.GetByLabel(d.user.Username, types.Label(label))
    if err != nil {
        fmt.Println("No entry found for label:", label)
        return
    }
    pretty, _ := json.MarshalIndent(data, "", "  ")
    fmt.Println(string(pretty))
}

func (d *Dashboard) removeByLabel(label string) {
    err := d.storage.RemoveFromStore(d.user.Username, types.Label(label))
    if err != nil {
        fmt.Println("Failed to remove:", err)
        return
    }
    fmt.Println("Entry removed successfully.")
}

func maskPassword(pwd string) string {
    if len(pwd) <= 4 {
        return "****"
    }
    return pwd[:2] + "****" + pwd[len(pwd)-2:]
}


func (d *Dashboard) wipeStore() {
	dir, err := utils.GetSessionDir()
	if err != nil {
		utils.ErrorE(err)
		return
	}

	path := filepath.Join(dir, consts.STORE_FILE_PATH)

	// Close DB if open
	if d.storage != nil {
		_ = d.storage.Close()
	}

	// Remove BadgerDB directory
	if err := os.RemoveAll(path); err != nil {
		utils.Error("Failed to wipe store: " + err.Error())
		return
	}

	// Reinitialize empty store
	newStore, err := store.NewStore()
	if err != nil {
		utils.Error("Failed to reinitialize store: " + err.Error())
		return
	}
	d.storage = newStore

	utils.Done("Store wiped successfully.")
}


func (d *Dashboard) getCurUser() {
	userSafe := struct {
		Username  string    `json:"username"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}{
		Username:  d.user.Username,
		CreatedAt: d.user.CreatedAt,
		UpdatedAt: d.user.UpdatedAt,
	}

	pretty, _ := json.MarshalIndent(userSafe, "", "  ")
	fmt.Println(string(pretty))
}

func (d *Dashboard) clearLog() {
	dir, err := utils.GetSessionDir()
	if err != nil {
		utils.ErrorE(err)
		return
	}

	filePath := filepath.Join(dir, consts.ROOK_LOG)
	err = os.Remove(filePath)
	if err != nil {
		utils.Error("Error deleting file: " + err.Error())
		return
	}

	utils.Done("Log cleared.")
}

func (d *Dashboard) viewLogs() {
	dir, err := utils.GetSessionDir()
	if err != nil {
		utils.ErrorE(err)
		return
	}

	logPath := filepath.Join(dir, consts.ROOK_LOG)
	content, err := os.ReadFile(logPath)
	if err != nil {
		utils.Error("Failed to read log file: " + err.Error())
		return
	}

	lines := strings.Split(string(content), "\n")
	count := 50
	if len(lines) < count {
		count = len(lines)
	}
	d.printRecentLogs(lines, count)
}


func cmdList() {
    width := 50 // Default width, you could detect terminal width dynamically

    top := "╭" + strings.Repeat("─", width-2) + "╮"
    sep := "├" + strings.Repeat("─", width-2) + "┤"
    bot := "╰" + strings.Repeat("─", width-2) + "╯"

    fmt.Println(top)
    fmt.Printf("│%-*s│\n", width-2, centerText("\033[1;36mROOK DASHBOARD\033[0m", width-2))
    fmt.Println(sep)

    fmt.Printf("│ \033[32m-l\033[0m %-*s│\n", width-6, "List all entries")
    fmt.Printf("│ \033[32m-a\033[0m %-*s│\n", width-6, "Add new entry")
    fmt.Printf("│ \033[32m-g\033[0m %-*s│\n", width-6, "Get entry by label")
    fmt.Printf("│ \033[32m-r\033[0m %-*s│\n", width-6, "Remove entry by label")
    fmt.Printf("│ \033[32m-w\033[0m %-*s│\n", width-6, "Wipe entire store")
    fmt.Printf("│ \033[32m-h\033[0m %-*s│\n", width-6, "Show this help")
    fmt.Printf("│ \033[32m-q\033[0m %-*s│\n", width-6, "Quit dashboard")
    fmt.Printf("│ \033[32m-q\033[0m %-*s│\n", width-6, "Clear App Logs")
    fmt.Printf("│ \033[32m-q\033[0m %-*s│\n", width-6, "View  App Logs")
    fmt.Printf("│ \033[32m-q\033[0m %-*s│\n", width-6, "User current user")

    fmt.Println(bot)
    fmt.Print("\033[34m»\033[0m ")
}

func centerText(text string, width int) string {
    // Removing ANSI escape sequences for length calculation
    plainText := removeANSICodes(text)
    textLen := len(plainText)

    if textLen >= width {
        return text
    }

    leftPad := (width - textLen) / 2
    return strings.Repeat(" ", leftPad) + text
}

func removeANSICodes(str string) string {
    // Simple function to remove ANSI escape sequences for length calculation
    // A regex could be used, but this simple approach suffices here
    result := ""
    inEscape := false
    for i := 0; i < len(str); i++ {
        if str[i] == 0x1b { // ESC character
            inEscape = true
            continue
        }
        if inEscape {
            if (str[i] >= 'a' && str[i] <= 'z') || (str[i] >= 'A' && str[i] <= 'Z') {
                inEscape = false
            }
            continue
        }
        result += string(str[i])
    }
    return result
}

func (d *Dashboard) printRecentLogs(lines []string, count int) {
    if len(lines) == 0 {
        fmt.Println("(No Logs Available)")
        return
    }

    fmt.Println("<<<< LAST LOGS >>>>")
    start := 0
    if len(lines) > count {
        start = len(lines) - count
    }
    for _, line := range lines[start:] {
        fmt.Println(line)
    }
    fmt.Println("<<<< END LOGS >>>>")
}