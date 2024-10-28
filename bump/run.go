package bump

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/joe-at-startupmedia/version-bump/v2/console"
	"golang.org/x/mod/semver"
)

var (
	ErrStrResponseHasEmptyTag = "tag name from response was empty"

	ErrStrFormattedUnsuccessfulStatusCode = "status code was not success: %d"
)

func init() {
	cli := &http.Client{
		Timeout: time.Second * 5,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 3 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	ReleaseGetter = cli
}

func (b *Bump) Run(ra *RunArgs) error {
	// check for an update in parallel
	updateVersion := make(chan string, 1)
	updateVersionError := make(chan error, 1)

	go getLatestVersion(updateVersion, updateVersionError, GhRepoName)

	if err := b.Bump(ra); err != nil {
		return err
	}

	err := <-updateVersionError

	if err != nil {
		console.ErrorCheckingForUpdate(err)
		return err
	}

	v := <-updateVersion
	if v != "" {
		console.UpdateAvailable(v, GhRepoName)
	}

	return err
}

func getLatestVersion(version chan string, resultErr chan error, repoName string) {

	type response struct {
		TagName string `json:"tag_name"`
	}

	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repoName)

	res, err := ReleaseGetter.Get(apiUrl)
	if err != nil {
		resultErr <- err
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		resultErr <- fmt.Errorf(ErrStrFormattedUnsuccessfulStatusCode, res.StatusCode)

	} else {
		d := new(response)
		if err = json.NewDecoder(res.Body).Decode(d); err != nil {
			resultErr <- err
			return
		}
		if d.TagName == "" {
			resultErr <- fmt.Errorf(ErrStrResponseHasEmptyTag)
		} else if semver.Compare(d.TagName, fmt.Sprintf("v%v", Version)) == 1 {
			resultErr <- nil
			version <- d.TagName
			return
		}
	}
	resultErr <- nil
	version <- "" // no new updates
}
