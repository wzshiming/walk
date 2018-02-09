package walk

import (
	"os"
	"os/exec"
	"path"
	"strings"
)

func getPath() ([]string, error) {
	gopath := []string{}
	goEnvGopath, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return nil, err
	}

	for _, v := range strings.Split(strings.TrimSpace(string(goEnvGopath)), ";") {
		gopath = append(gopath, v)
	}

	goEnvGoroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		return nil, err
	}

	gopath = append(gopath, strings.TrimSpace(string(goEnvGoroot)))

	for i := 0; i != len(gopath); {
		gopath[i] = path.Join(path.Clean(gopath[i]), "src")
		fi, err := os.Stat(gopath[i])
		if err != nil || !fi.IsDir() {
			gopath = append(gopath[:i], gopath[i+1:]...)
			continue
		}
		i++
	}
	return gopath, nil
}
