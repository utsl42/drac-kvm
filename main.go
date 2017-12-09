// -*- go -*-

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
	"io/ioutil"

	"github.com/haad/drac-kvm/kvm"

	"github.com/Unknwon/goconfig"
	"github.com/howeyc/gopass"
	"github.com/ogier/pflag"
)

// CLI flags
var _host = pflag.StringP("host", "h", "some.hostname.com", "The DRAC host (or IP)")
var _username = pflag.StringP("username", "u", "", "The DRAC username")
var _password = pflag.BoolP("password", "p", false, "Prompt for password (optional, will use 'calvin' if not present)")
var _version = pflag.IntP("version", "v", -1, "iDRAC version (6, 7 or 8)")
var _delay = pflag.IntP("delay", "d", 10, "Number of seconds to delay for javaws to start up & read jnlp before deleting it")
var _javaws = pflag.StringP("javaws", "j", DefaultJavaPath(), "The path to javaws binary")
var _wait = pflag.BoolP("wait", "w", false, "Wait for java console process end")

func promptPassword() string {
	fmt.Print("Password: ")
	password, _ := gopass.GetPasswd()
	return string(password)
}

func getJavawsArgs(waitFlag bool) string {
	var javawsArgs = "-jnlp"

	cmd := exec.Command("java", "-version")
	stderr, err := cmd.StderrPipe()

	if err != nil {
		//os.Remove(filename)
		log.Fatalf("Java not present on your system...", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	slurp, _ := ioutil.ReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	if strings.Contains(string(slurp[:]), "1.7") ||
		strings.Contains(string(slurp[:]), "1.8") {
		if waitFlag {
			javawsArgs = "-wait"
		} else {
			javawsArgs = ""
		}

	}

	return javawsArgs
}

func main() {
	var host string
	var username string
	var password string
	var version int

	var vendor = "dell"

	// Parse the CLI flags
	pflag.Parse()

	// Check we have access to the javaws binary
	if _, err := os.Stat(*_javaws); err != nil {
		log.Fatalf("No javaws binary found at %s", *_javaws)
	}

	// Search for existing config file
	usr, _ := user.Current()
	cfg, _ := goconfig.LoadConfigFile(usr.HomeDir + "/.drackvmrc")

	// Get the default username and password from the config
	if cfg != nil {
		_, err := cfg.GetSection("defaults")
		if err == nil {
			log.Printf("Loading default username and password from configuration file")
			uservalue, uerr := cfg.GetValue("defaults", "username")
			passvalue, perr := cfg.GetValue("defaults", "password")

			if uerr == nil {
				username = uservalue
			} else {
				username = kvm.GetDefaultUsername(vendor)
			}
			if perr == nil {
				password = passvalue
			} else {
				password = kvm.GetDefaultPassword(vendor)
			}
		}
	}

	// Finding host in config file or using the one passed in param
	host = *_host
	hostFound := false
	if cfg != nil {
		_, err := cfg.GetSection(*_host)
		if err == nil {
			value, err := cfg.GetValue(*_host, "host")
			if err == nil {
				hostFound = true
				host = value
			} else {
				hostFound = true
				host = *_host
			}
		}
	}

	if *_username != "" {
		username = *_username
	} else {
		if cfg != nil && hostFound {
			value, err := cfg.GetValue(*_host, "username")
			if err == nil {
				username = value
			}
		}
	}

	// If password not set, prompt
	if *_password {
		password = promptPassword()
	} else {
		if cfg != nil && hostFound {
			value, err := cfg.GetValue(*_host, "password")
			if err == nil {
				password = value
			}
		}
	}

	version = *_version
	if cfg != nil && hostFound {
		value, err := cfg.Int(*_host, "version")
		if err == nil {
			version = value
		}
	}

	if username == "" && password == "" {
		log.Printf("Username/Password not provided trying without them...")
	}

	filename := kvm.CreateKVM(host, username, password, vendor, version, true).GetJnlpFile()
	defer os.Remove(filename)

	// Launch it!
	log.Printf("Launching KVM session with %s", filename)
	if err := exec.Command(*_javaws, getJavawsArgs(*_wait), filename, "-nosecurity", "-noupdate", "-Xnofork").Run(); err != nil {
		os.Remove(filename)
		log.Fatalf("Unable to launch DRAC (%s), from file %s", err, filename)
	}

	// Give javaws a few seconds to start & read the jnlp
	time.Sleep(time.Duration(*_delay) * time.Second)
}

// EOF
