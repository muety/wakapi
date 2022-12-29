package models

import (
	"os"
	"path"
	"runtime"
)

func init() {
	// move to project root as working directory (e.g. so data/* can be resolved when loading config)
	// taken from https://intellij-support.jetbrains.com/hc/en-us/community/posts/360009685279-Go-test-working-directory-keeps-changing-to-dir-of-the-test-file-instead-of-value-in-template
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
