package constants

import (
	"log"
	"os"
	"runtime"
)

// ChromePath is the path of chrome used for run html2image server
var ChromePath string

// UserDataDir is the chrome workpath
var UserDataDir string

func init() {
	defaultChromePathDarwin := `/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary`
	defaultChromePathLinux := `/usr/bin/google-chrome-unstable`

	var ok bool
	chromePath, ok := os.LookupEnv("CHROME_PATH")
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
	log.Println(chromePath)
	ChromePath = chromePath
}
