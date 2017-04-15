package scrape

import (
	"os/exec"
)


// Uses phantomjs to load a site and let the js execute before returning
// the resulting html.
func ReadSite(url string) (html []byte, err error) {
	html, err = exec.Command("phantomjs", "content.js", url).Output()
	return
}
