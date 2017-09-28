// -*- go -*-

package main

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"
	"encoding/json"
	"io/ioutil"
)

// DRAC contains all of the information required
// to connect to a Dell DRAC KVM
type DRAC struct {
	Host     string
	Username string
	Password string
	Version  int
}

// Templates is a map of each viewer.jnlp template for
// the various Dell iDRAC versions, keyed by version number
var Templates = map[int]string{
	6: viewer6,
	7: viewer7,
	8: viewer8,
	103: viewer7,
	104: viewer7,
}

// GetVersion attempts to detect the iDRAC version by checking
// if various known libraries are available via HTTP GET requests.
// Retursn the version if found, or -1 if unknown
func (d *DRAC) GetVersion() int {

	log.Print("Detecting iDRAC version...")

	version := -1

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
			return 103
		}
	}

	// Check for iLO4 specific libs
	if response, err := client.Get("https://" + d.Host + "/html/intgapp4_231.jar"); err == nil {
		defer response.Body.Close()
		if response.StatusCode == 200 {
			return 104
		}
	}

	// Check for iDRAC6 specific libs
	if response, err := client.Head("https://" + d.Host + "/software/jpcsc.jar"); err == nil {
		defer response.Body.Close()
		if response.StatusCode == 200 {
			return 6
		}
	}

	// Check for iDRAC7 specific libs
	if response, err := client.Head("https://" + d.Host + "/software/avctKVMIOMac64.jar"); err == nil {
		defer response.Body.Close()
		if response.StatusCode == 200 {
			return 7
		}
	}

	// Check for iDRAC8 specific libs
	// if response, err := client.Head("https://" + d.Host + "/images/Ttl_2_iDRAC8_Base_ML.png"); err == nil {
	// 	defer response.Body.Close()
	// 	if response.StatusCode == 200 {
	// 		return 8
	// 	}
	// }

	return version

}

// iLO Viewer that logs in, fetch the sessionKey cookie to be able to generate
// a correct jnlp
func (d *DRAC) iLOViewer() (string, error) {
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

	// Post parameters to login to iLO
	values := map[string]string{"method": "login", "user_login": d.Username, "password": d.Password}
	jsonValue, _ := json.Marshal(values)

	if res, err := client.Post("https://" + d.Host + "/json/login_session", "", bytes.NewBuffer(jsonValue)); err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			// Fetch sessionKey from json response using an
			// interface in order to build a cookie
			bodyBytes, err := ioutil.ReadAll(res.Body)
			var f interface{}
			err = json.Unmarshal(bodyBytes, &f)
			if err != nil {
				fmt.Println("Couldn't decode json", err)
			}
			m := f.(map[string]interface{})
			sessionKey := m["sessionKey"].(string)

			cookie := http.Cookie{Name: "sessionKey", Value: sessionKey}
			req, err := http.NewRequest("GET", "https://" + d.Host + "/html/jnlp_template.html", nil)
			req.AddCookie(&cookie)
			if res, err := client.Do(req); err == nil {
				defer res.Body.Close()
				if res.StatusCode == 200 {
					bodyBytes, _ := ioutil.ReadAll(res.Body)
					bodyString := string(bodyBytes)
					// We need to:
					// - replace placeholder with actual values
					// - skip the first and last line of the jnlp
					// template
					r := strings.NewReplacer("<%= this.baseUrl %>", "https://" + d.Host + "/",
						"<%= this.sessionKey %>", sessionKey,
						"<%= this.langId %>", "en")
					jnlp := r.Replace(bodyString)
					_jnlp := strings.Split(jnlp, "\n")
					jnlp = strings.Join(_jnlp[1 : len(_jnlp) - 1],"\n")
					return jnlp, err
				} else {
					return "", errors.New("Couldn't fetch jnlp template")
				}
			}
		}
	}
	return "", errors.New("Couldn't login to iLO")
}

// Viewer returns a viewer.jnlp template filled out with the
// necessary details to connect to a particular DRAC host
func (d *DRAC) Viewer() (string, error) {

	var version int

	// Check we have a valid DRAC viewer template for this DRAC version
	if d.Version < 0 {
		version = d.GetVersion()
	} else {
		version = d.Version
	}
	if version < 0 {
		return "", errors.New("unable to detect DRAC version")
	}


	if _, ok := Templates[version]; !ok {
		msg := fmt.Sprintf("no support for DRAC v%d", version)
		return "", errors.New(msg)
	}

	// If it's an iLO version
	if version > 100 {
		log.Printf("Found iLO version %d", version - 100)
		return d.iLOViewer()
	} else {
		log.Printf("Found iDRAC version %d", version)
		// Generate a JNLP viewer from the template
		// Injecting the host/user/pass information
		buff := bytes.NewBufferString("")
		err := template.Must(template.New("viewer").Parse(Templates[version])).Execute(buff, d)
		return buff.String(), err
	}
}
// EOF
