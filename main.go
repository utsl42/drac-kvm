// -*- go -*-

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/utsl42/drac-kvm/kvm"

	"github.com/Unknwon/goconfig"
	"github.com/howeyc/gopass"
	"github.com/ogier/pflag"
)

const (
	// DracKVMVersion current application version
	DracKVMVersion = "2.0.1"
)

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
		log.Fatalf("Java not present on your system... (%s)", err)
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
	var vendor string
	var username string
	var password string
	var version int

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Program %s version: %s\n\n", os.Args[0], DracKVMVersion)
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
	}

	// CLI flags
	var _host = pflag.StringP("host", "h", "", "The DRAC host (or IP)")
	var _vendor = pflag.StringP("vendor", "V", "", "The KVM Vendor")

	var _username = pflag.StringP("username", "u", "", "The KVM username")
	var _password = pflag.BoolP("password", "p", false, "Prompt for password (optional, will use default vendor if not present)")
	var _version = pflag.IntP("version", "v", -1, "KVM vendor specific version for idrac: (6, 7 or 8)")

	var _delay = pflag.IntP("delay", "d", 10, "Number of seconds to delay for javaws to start up & read jnlp before deleting it")
	var _javaws = pflag.StringP("javaws", "j", DefaultJavaPath(), "The path to javaws binary")
	var _wait = pflag.BoolP("wait", "w", false, "Wait for java console process end")

	// Parse the CLI flags
	pflag.Parse()

	if *_host == "" {
		log.Printf("Host parameter is requried...")
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// Check we have access to the javaws binary
	if _, err := os.Stat(*_javaws); err != nil {
		log.Fatalf("No javaws binary found at %s", *_javaws)
	}

	// Search for existing config file
	usr, _ := user.Current()
	cfg, _ := goconfig.LoadConfigFile(usr.HomeDir + "/.drackvmrc")

	/*
	 *	Values loaded from config file has lower priority than command line arguments.
	 *  For each possible option we first check if command line argument was passed and
	 *  if not then we try to get value from config file.
	 *
	 */
	if value, err := cfg.GetValue(*_host, "host"); err == nil {
		host = value
	} else {
		host = *_host
	}

	/*
	 *	For loading vendor string we have following order:
	 *
	 *	1) Check if vendor was used as command line argument
	 *	2) Try to load it from _host_ section of config
	 *	3) Check if _defaults_ section of config contains _vendor_
	 *	4) Use default "dell" value to keep original behaviour
	 *
	 */
	if *_vendor == "" {
		if value, err := cfg.GetValue(*_host, "vendor"); err == nil {
			vendor = value
		} else {
			// To keep old default behaviour we set vendor string to dell by default.
			vendor = "dell"
		}
	} else {
		vendor = *_vendor
	}

	if _, err := kvm.CheckVendorString(vendor); err != nil {
		log.Fatalf("Provided vendor: %s, is not supported consider adding support with Github PR...", vendor)
	}

	/*
	 *  For loading username/password we have following order:
	 *
	 *	1) Check if username/password was used as argument
	 *  2) Try to load them from _host_ section of config
	 *  3) Check if _defaults_ section of our config contains username/password
	 *  4) Use default vendor provided values defined in vendor packages.
	 */
	if *_username == "" {
		if value, err := cfg.GetValue(*_host, "username"); err == nil {
			username = value
		} else {
			if defaultvalue, err := cfg.GetValue("defaults", "username"); err == nil {
				username = defaultvalue
			} else {
				username = kvm.GetDefaultUsername(vendor)
			}
		}
	} else {
		username = *_username
	}

	if !*_password {
		if value, err := cfg.GetValue(*_host, "password"); err == nil {
			password = value
		} else {
			if defaultvalue, err := cfg.GetValue("defaults", "password"); err == nil {
				password = defaultvalue
			} else {
				password = kvm.GetDefaultPassword(vendor)
			}
		}
	} else {
		password = promptPassword()
	}

	// Version is only used with dell KVM vendor..
	if vendor == "dell" && *_version == -1 {
		if value, err := cfg.Int(*_host, "version"); err == nil {
			version = value
		} else {
			if defaultvalue, err := cfg.Int("defaults", "version"); err == nil {
				version = defaultvalue
			}
		}
	} else {
		version = *_version
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
