// -*- go -*-

package supermicro

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"text/template"
)

// KvmSupermicroDriver is Supermicro specific folder for KVM driver.
//
type KvmSupermicroDriver struct {
	Host     string
	Username string
	Password string
	Version  int
}

const (
	// DefaultUsername is the default username on Supermicro KVM
	DefaultUsername = "ADMIN"
	// DefaultPassword is the default password on Supermicro KVM
	DefaultPassword = "ADMIN"
)

// SupermicroTemplates is a map of each viewer.jnlp template for
// the various Supermicro iKVM versions, keyed by version number
var SupermicroTemplates = map[int]string{
	169:   ikvm169,
}

// Viewer returns a viewer.jnlp template filled out with the
// necessary details to connect to a particular DRAC host
func (d *KvmSupermicroDriver) Viewer() (string, error) {

		var version int

		if _, ok := SupermicroTemplates[version]; !ok {
			msg := fmt.Sprintf("no support for iKVM v%d", version)
			return "", errors.New(msg)
		}

		log.Printf("Found iKVM version %d", version)
		// Generate a JNLP viewer from the template
		// Injecting the host/user/pass information
		buff := bytes.NewBufferString("")
		err := template.Must(template.New("viewer").Parse(SupermicroTemplates[version])).Execute(buff, d)
		return buff.String(), err
}

// GetHost return Configured driver Host
func (d *KvmSupermicroDriver) GetHost() string {
	return d.Host
}

// GetUsername return Configured driver Username
func (d *KvmSupermicroDriver) GetUsername() string {
	return d.Username
}

// GetPassword return Configured driver Password
func (d *KvmSupermicroDriver) GetPassword() string {
	return d.Password
}

// EOF
