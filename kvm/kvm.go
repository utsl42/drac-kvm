// -*- go -*-

package kvm

import (
	"github.com/haad/drac-kvm/dell"
	"github.com/haad/drac-kvm/hp"
	"github.com/haad/drac-kvm/supermicro"
	"io/ioutil"
	"log"
	"os"
)

// Driver is interface for all usable kvm drivers
// 	HP iLO, Dell idrac, Supermicro, IBM
// Every one of them needs to support following methods
//  - Viewer which will return buffer with generated template
//  - GetHost/GetUsername/GetPassword
type Driver interface {
	Viewer() (string, error)
	GetHost() string
	GetUsername() string
	GetPassword() string
}

// Config is simple config structure
type Config struct {
	InsecureSkipVerify bool
}

// KVM contains all of the information required
// to connect to a KVM
type KVM struct {
	Vendor   string
	Config
	Driver
}

// CreateKVM will create KVM structure based on input it will assign proper
// driver to interface.
func CreateKVM(Host string, Username string, Password string, Vendor string,
	Version int, InsecureSkipVerify bool) *KVM {

	var driver Driver

	switch vn := Vendor; vn {
		case "dell":
			driver = &dell.KvmDellDriver{
				Host: 	  Host,
				Username: Username,
				Password: Password,
				Version:  Version,
			}
		case "supermicro":
			driver = &supermicro.KvmSupermicroDriver{
				Host: 	  Host,
				Username: Username,
				Password: Password,
				Version:  169,
			}
		case "hp":
			driver = &hp.KvmHpDriver{
				Host: 	  Host,
				Username: Username,
				Password: Password,
				Version:  -1,
			}
		case "ibm":
			log.Fatalf("IBM/Lennovo support not implemented yet KVM as driver: %s", vn)
		default:
			log.Fatalf("Unsupported KVM vendor %s", vn)
	}

	kvm := &KVM{
		Vendor:   Vendor,
		Config: Config{
			InsecureSkipVerify: InsecureSkipVerify,
		},
		Driver: driver,
	}

	return kvm
}

/* getVersion attempts to detect the version of iKVM/iLO/by checking
 if various known libraries are available via HTTP GET requests.
 Returns the version if found, or -1 if unknown
 TODO: Figure out how to detect supermicro/iLO if possible.
func (d *KVM) getVersion() (string, int) {

		log.Print("Detecting KVM vendor and driver version...")

		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(5 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*5)
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		// Check for iLO3 specific libs
		if response, err := client.Get("https://" + d.Host + "/html/intgapp3_231.jar"); err == nil {
			defer response.Body.Close()
			if response.StatusCode == 200 {
				return "dell", 103
			}
		}

		// Check for iLO4 specific libs
		if response, err := client.Get("https://" + d.Host + "/html/intgapp4_231.jar"); err == nil {
			defer response.Body.Close()
			if response.StatusCode == 200 {
				return "dell", 104
			}
		}

		// Check for iDRAC6 specific libs
		if response, err := client.Head("https://" + d.Host + "/software/jpcsc.jar"); err == nil {
			defer response.Body.Close()
			if response.StatusCode == 200 {
				return "dell", 6
			}
		}

		// Check for iDRAC7 specific libs
		if response, err := client.Head("https://" + d.Host + "/software/avctKVMIOMac64.jar"); err == nil {
			defer response.Body.Close()
			if response.StatusCode == 200 {
				return "dell", 7
			}
		}

		return -1
	}
*/

// GetJnlpFile Creates JNLP file and return PATH to it
func (d *KVM) GetJnlpFile() string {

	viewer, err := d.Driver.Viewer()
	if err != nil {
		log.Fatalf("Unable to generate DRAC viewer for %s@%s (%s)", d.Driver.GetUsername(), d.Driver.GetHost(), err)
	}

	// Write out the kvm viewer to a temporary file so that
	// we can launch it with the javaws program
	filename := os.TempDir() + string(os.PathSeparator) + "kvm_" + d.Driver.GetHost() + ".jnlp"

	ioutil.WriteFile(filename, []byte(viewer), 0600)

	return filename
}

// GetDefaultUsername returns default KVM vendor user
func GetDefaultUsername(Vendor string) string {
	switch vn := Vendor; vn {
	case "dell":
		return dell.DefaultUsername
	case "supermicro":
		return supermicro.DefaultUsername
	case "hp":
		return hp.DefaultUsername
	case "IBM":
		log.Fatalf("IBM/Lennovo support not implemented yet KVM vendor %s", vn)
	default:
		log.Fatalf("Unsupported KVM vendor %s", vn)
	}
	return ""
}

// GetDefaultPassword returns default KVM vendor password
func GetDefaultPassword(Vendor string) string {
	switch vn := Vendor; vn {
	case "dell":
		return dell.DefaultPassword
	case "supermicro":
		return supermicro.DefaultPassword
	case "hp":
		return hp.DefaultPassword
	case "IBM":
		log.Fatalf("IBM/Lennovo support not implemented yet KVM vendor %s", vn)
	default:
		log.Fatalf("Unsupported KVM vendor %s", vn)
	}
	return ""
}

// EOF
