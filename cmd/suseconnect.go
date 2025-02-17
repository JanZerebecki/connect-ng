package main

import (
	_ "embed"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/SUSE/connect-ng/connect"
)

var (
	//go:embed usage.txt
	usageText   string
	status      bool
	statusText  bool
	debug       bool
	writeConfig bool
	deregister  bool
	baseURL     string
	fsRoot      string
	namespace   string
)

func init() {
	// display help like the ruby SUSEConnect
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
	}

	flag.BoolVar(&status, "status", false, "")
	flag.BoolVar(&status, "s", false, "")
	flag.BoolVar(&statusText, "status-text", false, "")
	flag.BoolVar(&debug, "debug", false, "")
	flag.BoolVar(&writeConfig, "write-config", false, "")
	flag.BoolVar(&deregister, "deregister", false, "")
	flag.BoolVar(&deregister, "d", false, "")
	flag.StringVar(&baseURL, "url", "", "")
	flag.StringVar(&fsRoot, "root", "", "")
	flag.StringVar(&namespace, "namespace", "", "")
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "Root privileges are required to register products and change software repositories.")
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	flag.Parse()
	if debug {
		connect.EnableDebug()
	}
	connect.Debug.Println("cmd line:", os.Args)
	connect.Debug.Println("For http debug use: GODEBUG=http2debug=2", strings.Join(os.Args, " "))
	connect.CFG.Load()
	if baseURL != "" {
		if err := validateURL(baseURL); err != nil {
			fmt.Printf("URL \"%s\" not valid: %s\n", baseURL, err)
			os.Exit(1)
		}
		connect.CFG.BaseURL = baseURL
		writeConfig = true
	}
	if fsRoot != "" {
		if fsRoot[0] != '/' {
			fmt.Println("The path specified in the --root option must be absolute.")
			os.Exit(1)
		}
		connect.CFG.FsRoot = fsRoot
	}
	if namespace != "" {
		connect.CFG.Namespace = namespace
		writeConfig = true
	}
	if lang, ok := os.LookupEnv("LANG"); ok {
		if lang != "" {
			connect.CFG.Language = lang
		}
	}
	if status {
		output, err := connect.GetProductStatuses("json")
		exitOnError(err)
		fmt.Println(output)
	} else if statusText {
		output, err := connect.GetProductStatuses("text")
		exitOnError(err)
		fmt.Print(output)
	} else if deregister {
		err := connect.Deregister()
		exitOnError(err)
	}
	if writeConfig {
		if err := connect.CFG.Save(); err != nil {
			fmt.Printf("Problem writing configuration: %s\n", err)
			os.Exit(1)
		}
	}
}

func exitOnError(err error) {
	if err == nil {
		return
	}
	if ze, ok := err.(connect.ZypperError); ok {
		fmt.Printf("%s\n", ze)
		os.Exit(ze.ExitCode)
	}
	if ae, ok := err.(connect.APIError); ok {
		if ae.Code == http.StatusUnauthorized && connect.IsRegistered() {
			fmt.Print("Error: Invalid system credentials, probably because the ")
			fmt.Print("registered system was deleted in SUSE Customer Center. ")
			fmt.Print("Check ", connect.CFG.BaseURL, " whether your system appears there. ")
			fmt.Print("If it does not, please call SUSEConnect --cleanup and re-register this system.\n")
		} else if !connect.URLDefault() && !connect.UpToDate() {
			fmt.Print("Your Registration Proxy server doesn't support this function. ")
			fmt.Print("Please update it and try again.")
		} else {
			fmt.Printf("%s\n", ae)
		}
		os.Exit(67)
	}
	switch err {
	case connect.ErrSystemNotRegistered:
		fmt.Print("Deregistration failed. Check if the system has been ")
		fmt.Print("registered using the --status-text option or use the ")
		fmt.Print("--regcode parameter to register it.\n")
		os.Exit(69)
	default:
		fmt.Printf("Command exited with error: %s\n", err)
		os.Exit(1)
	}
}

func validateURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	if u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("Missing scheme or host")
	}
	return nil
}
