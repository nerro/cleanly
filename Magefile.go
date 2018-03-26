// +build mage

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	buildDir     = "build"
	pkgPrefixLen = len("github.com/nerro/cleanly")
)

// Delete the build directory.
func Clean() error {
	fmt.Printf("==> Removing '%v' directory...\n", buildDir)

	if err := os.RemoveAll(buildDir); err != nil {
		return err
	}
	return nil
}

// Build release binaries for all supported versions.
//
// Currently that means a combination of windows, linux and OSX in 32 bit and 64 bit formats.
// The files will be dumped into 'build' directory. Finally sha256 checksum for all files will
// be generated.
//
// Depends on 'fmt' task.
func Release() error {
	mg.Deps(Fmt)
	fmt.Println("==> Building release binaries...")

	for _, goos := range []string{"windows", "linux", "darwin"} {
		for _, goarch := range []string{"amd64", "386"} {
			if err := build(goos, goarch); err != nil {
				return err
			}
		}
	}

	files, err := ioutil.ReadDir(buildDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := writeHashSumFile(f.Name()); err != nil {
			return err
		}
	}
	return nil
}

// Build binary for default local system's operating system and architecture.
//
// Depends on 'fmt' task.
func Build() error {
	mg.Deps(Fmt)
	fmt.Println("==> Building binary...")

	return build(runtime.GOOS, runtime.GOARCH)
}

// Run gofmt linter.
func Fmt() error {
	fmt.Println("==> Checking code with 'go fmt'...")

	pkgs, err := cleanlyPackages()
	if err != nil {
		return err
	}
	first := true
	for _, pkg := range pkgs {
		files, err := filepath.Glob(filepath.Join(pkg, "*.go"))
		if err != nil {
			return nil
		}
		for _, f := range files {
			gofmtOutput, err := sh.Output("gofmt", "-l", f)
			if err != nil {
				fmt.Printf("ERROR: running gofmt no %q: %v\n", f, err)
			}
			if gofmtOutput != "" {
				if first {
					fmt.Println("    following files are not gofmt'ed:")
					first = false
				}
				fmt.Printf("    - %v\n", gofmtOutput)
			}
		}
	}
	return nil
}

func cleanlyPackages() ([]string, error) {
	s, err := sh.Output("go", "list", "./...")
	if err != nil {
		return nil, err
	}
	pkgs := strings.Split(s, "\n")
	for i := range pkgs {
		pkgs[i] = "." + pkgs[i][pkgPrefixLen:]
	}
	return pkgs, nil
}

func build(goos, goarch string) error {
	fmt.Printf("    running go build for GOOS=%v GOARCH=%v\n", goos, goarch)

	env := map[string]string{"GOOS": goos, "GOARCH": goarch}
	filename := fmt.Sprintf("%s/cleanly_%v_%v", buildDir, goos, goarch)
	if goos == "windows" {
		filename = filename + ".exe"
	}

	if err := sh.RunWith(env, "go", "build", "-o", filename, "cmd/cleanly.go"); err != nil {
		return err
	}
	return nil
}

func writeHashSumFile(filename string) error {
	file, err := os.Open(buildDir + string(os.PathSeparator) + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	hashInBytes := hash.Sum(nil)

	checksumFile, err := os.Create(buildDir + string(os.PathSeparator) + filename + ".sha256")
	if err != nil {
		return err
	}
	defer checksumFile.Close()

	fmt.Fprintf(checksumFile, "%v *%v", hex.EncodeToString(hashInBytes), filename)
	return nil
}
