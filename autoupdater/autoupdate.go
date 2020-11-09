package autoupdater

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type AutoUpdater struct {
	Applications []*Application
}

func (au *AutoUpdater) ParseConfiguration(config string) (apps []string, err error) {

	configBytes, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration: %v", err)
	}

	au.Applications = make([]*Application, 0)
	err = yaml.Unmarshal(configBytes, &au.Applications)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration %v", err)
	}

	var names []string
	for _, app := range au.Applications {
		names = append(names, app.Name)
	}
	return names, nil
}

func (au *AutoUpdater) Update(name string) (err error) {

	for _, app := range au.Applications {
		if app.Name == name {
			log.Printf("checking for update to %s", name)
			localVerB, err := ioutil.ReadFile(path.Join(app.LocalDir, app.VersionFile))
			if err != nil && os.IsNotExist(err) {
				// ignore we'll just assume we need to update the application if it doesn't exist
				if err := os.MkdirAll(app.LocalDir, 0755); err != nil {
					return fmt.Errorf("failed to create directory for app %s: %v", name, err)
				} else {
					f, err := os.Create(path.Join(app.LocalDir, app.VersionFile))
					if err != nil {
						return fmt.Errorf("failed to create new version file: %v", err)
					}
					f.WriteString("0")
					f.Close()
					localVerB = []byte("0")
				}
			} else if err != nil {
				return fmt.Errorf("failed to read local version")
			}

			localVer := strings.Replace(string(localVerB), "\n", "", 1)
			localVerNum, err := strconv.Atoi(localVer)
			if err != nil {
				return fmt.Errorf("failed to parse local version: %v", err)
			}

			resp, err := http.Get(fmt.Sprintf("%s/%s", app.UpdateUrl, app.VersionFile))
			if err != nil {
				return fmt.Errorf("failed to retrieve remote version: %v", err)
			}
			defer resp.Body.Close()

			remoteVerB, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read remote version: %v", err)
			}
			remoteVer := strings.Replace(string(remoteVerB), "\n", "", 1)
			remoteVerNum, err := strconv.Atoi(remoteVer)
			if err != nil {
				return fmt.Errorf("failed to parse remote version: %v", err)
			}

			if localVerNum < remoteVerNum {
				log.Printf(" application update needed. local: %s remote: %s\n", localVer, remoteVer)

				// download the tar to a temp directory
				tempDir, err := ioutil.TempDir("", name)
				if err != nil {
					return err
				}

				remoteTar := fmt.Sprintf("%s/%s", app.UpdateUrl, app.PackageName)
				tempTar := path.Join(tempDir, app.PackageName)

				err = downloadFile(tempTar, remoteTar)
				if err != nil {
					return err
				}

				// stop the app
				_, err = ExecuteCommand(app.StopCmd)

				_, err = ExecuteCommand(fmt.Sprintf("rm -rf %s/*", app.LocalDir))
				_, err = ExecuteCommand(fmt.Sprintf("tar -xf %s -C %s", tempTar, app.LocalDir))
				if err != nil {
					return err
				}
				err = ioutil.WriteFile(path.Join(app.LocalDir, app.VersionFile), []byte(remoteVer), 0755)
				if err != nil {
					return fmt.Errorf("failed to write updated version file: %v", err)
				}

				// start the app
				_, err = ExecuteCommand(app.StartCmd)

				log.Printf(" finished update")

			} else {
				log.Printf(" no update needed at this time. local: %s remote: %s\n", localVer, remoteVer)
				return nil
			}

			return nil
		}
	}

	return fmt.Errorf("unable to find applicatio update package for application: %s", name)

}

func downloadFile(localPath string, url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	outfile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer outfile.Close()

	_, err = io.Copy(outfile, resp.Body)
	return err
}
