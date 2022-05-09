package runtime

import (
	"runtime"
	"strings"
)

func GetPackageName(skip int) string {
	var packageName string
	if pc, _, _, ok := runtime.Caller(skip); ok {
		f := runtime.FuncForPC(pc)
		funcName := f.Name()
		lastSlash := strings.LastIndex(funcName, "/")
		if lastSlash > 0 {
			packageName = funcName[0:lastSlash]
			funcName = funcName[lastSlash:]
		}
		packageName += strings.SplitN(funcName, ".", 2)[0]
	}
	return packageName
}
