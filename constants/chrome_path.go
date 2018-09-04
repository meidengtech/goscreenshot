package constants

import (
	"log"
	"os"
	"runtime"
	"strconv"
)

// ChromePath is the path of chrome used for run html2image server
var ChromePath string

// UserDataDir is the chrome workpath
var UserDataDir string

// ServerPort is the http service port for this service
var ServerPort int

// DebugMode will open or close debugmode
var DebugMode bool

func init() {
	defaultChromePathDarwin := `/Applications/Google Chrome.app/Contents/MacOS/Google Chrome`
	defaultChromePathLinux := `/usr/bin/google-chrome-unstable`

	var ok bool
	chromePath, ok := os.LookupEnv("SCREENSHOT_CHROME_PATH")
	if !ok {
		if "linux" == runtime.GOOS {
			chromePath = defaultChromePathLinux
			UserDataDir = "/data"
		} else if "darwin" == runtime.GOOS {
			chromePath = defaultChromePathDarwin
			UserDataDir = "/tmp"
		} else {
			log.Fatal("os not support")
		}
	}
	if _, err := os.Stat(chromePath); err != nil {
		log.Fatalf("%s not exists", chromePath)
	}
	log.Println("GoogleChrome is: ", chromePath)
	ChromePath = chromePath

	ServerPort = 8080
	serverPort, ok := os.LookupEnv("SCREENSHOT_SERVER_PORT")
	if ok {
		var err error
		ServerPort, err = strconv.Atoi(serverPort)
		if err != nil {
			log.Fatalf("Parse HTTP Port Error")
		}
	}

	DebugMode = false
	debugMode, ok := os.LookupEnv("SCREENSHOT_DEBUG_MODE")
	if ok && debugMode != "" {
		DebugMode = true
	}

}
