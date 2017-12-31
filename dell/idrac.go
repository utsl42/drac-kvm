// -*- go -*-

package dell

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"text/template"
)

// KvmDellDriver is Dell specific folder for KVM driver.
//
type KvmDellDriver struct {
	Host     string
	Username string
	Password string
	Version  int
}

const (
	// DefaultUsername is the default username on Dell iDRAC
	DefaultUsername = "root"
	// DefaultPassword is the default password on Dell iDRAC
	DefaultPassword = "calvin"
)

// DellTemplates is a map of each viewer.jnlp template for
// the various Dell iDRAC versions, keyed by version number
var DellTemplates = map[int]string{
	6:   viewer6,
	7:   viewer7,
	8:   viewer8,
	103: viewer7,
	104: viewer7,
}

// Viewer returns a viewer.jnlp template filled out with the
// necessary details to connect to a particular DRAC host
func (d *KvmDellDriver) Viewer() (string, error) {

	// Check we have a valid DRAC viewer template for this DRAC version
	if d.Version < 0 {
		return "", errors.New("unable to detect DRAC version")
	}

	if _, ok := DellTemplates[d.Version]; !ok {
		msg := fmt.Sprintf("no support for DRAC v%d", d.Version)
		return "", errors.New(msg)
	}

	log.Printf("Found iDRAC version %d", d.Version)

	// Generate a JNLP viewer from the template
	// Injecting the host/user/pass information
	buff := bytes.NewBufferString("")
	err := template.Must(template.New("viewer").Parse(DellTemplates[d.Version])).Execute(buff, d)
	return buff.String(), err
}

// GetHost return Configured driver Host
func (d *KvmDellDriver) GetHost() string {
	return d.Host
}

// GetUsername return Configured driver Username
func (d *KvmDellDriver) GetUsername() string {
	return d.Username
}

// GetPassword return Configured driver Password
func (d *KvmDellDriver) GetPassword() string {
	return d.Password
}

// EOF
