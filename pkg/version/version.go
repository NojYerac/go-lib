package version

import (
	"runtime"
)

func GetVersion() Version {
	return Version{
		Name:   serviceName,
		SemVer: semVer,
		GitSHA: gitSHA,
		GoVer:  runtime.Version(),
		Arch:   runtime.GOARCH,
		OS:     runtime.GOOS,
	}
}

type Version struct {
	Name   string `json:"serviceName"`
	SemVer string `json:"semVer"`
	GitSHA string `json:"gitSHA,omitempty"`
	GoVer  string `json:"goVer"`
	Arch   string `json:"arch"`
	OS     string `json:"os"`
}

// go build -ldflags "-X version.gitSHA=$(git rev-list -1 HEAD)"

var (
	serviceName = ""
	semVer      = "0.0.0"
	gitSHA      = ""
)

func SetServiceName(name string) {
	serviceName = name
}

func SetSemVer(ver string) {
	semVer = ver
}
