// See README.md
package main

/*
Copyright 2014 Graham King

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

For full license details see <http://www.gnu.org/licenses/>.
*/

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"sort"
	"strings"
)

const usage = `USAGE: deps <package> [-display deep|layers -lib -stdlib -short]
"deps" prints the internal dependencies of a Go package.

-display deep|layers  Display more / different information
 deep: print the dependencies of the dependencies, recursively.
 layers: display the dependency layers

-lib  Include libraries.
 By default deps ignores anything starting with github.com, bitbucket.org, etc,
 because those are libraries and you only care about your app. Add this flag
 to prevent this ignoring.

-stdlib  Include Go built-in packages.
 By default deps ignores Go standard library packages. Add this flag
 to prevent this ignoring.

-short  Trim the package you are analyzing off the front of dependencies.
 e.g.: github.com/coreos/etcd/config -> config.

<package> is a path exactly like you would use in your code in "import".
That package and all it's dependencies must be on findable (GOPATH or stdlib).
`

var (
	thirdPartyRoots = []string{
		"github.com",
		"bitbucket.org",
		"launchpad.net",
		"code.google.com",
	}
	display         = flag.String("display", "deps", "Display format: deep|layers")
	isHelp          = flag.Bool("h", false, "Display this help")
	isIncludeStdlib = flag.Bool("stdlib", false, "Include standard library packages")
	isIncludeLibs   = flag.Bool("lib", false, "Include third-party library packages")
	isShort         = flag.Bool("short", false, "Trim current package name from dependencies")
	rootPackage     string
	deps            map[string][]*build.Package
	numDeps         map[string]int
	layerPos        map[string]int
	lowestLayer     int
	progress        int
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	flag.CommandLine.Parse(os.Args[2:])
	rootPackage = os.Args[1]

	if *isHelp {
		fmt.Println(usage)
		os.Exit(1)
	}

	numDeps = make(map[string]int)
	deps = make(map[string][]*build.Package)
	layerPos = make(map[string]int)

	fmt.Println("Dependencies of", bold(rootPackage))
	pkg, err := build.Import(rootPackage, "", 0)
	if err != nil {
		log.Fatal(err)
	}

	analyze(pkg, 0)
	os.Stdout.Write([]byte("                                     \r"))
	os.Stdout.Sync()

	switch *display {
	case "layers":
		layerDisplay()
	case "deep":
		deepDepsDisplay(rootPackage, 0)
	default:
		depsDisplay(rootPackage)
	}
}

func analyze(pkg *build.Package, layer int) {
	os.Stdout.Write([]byte(fmt.Sprintf("Working ... %d   \r", progress)))
	progress++
	os.Stdout.Sync()

	if layer > lowestLayer {
		lowestLayer = layer
	}
	path := pkg.ImportPath
	if val, ok := layerPos[path]; ok && layer <= val {
		// We've already found this package at a deeper layer
		return
	}

	layerPos[path] = layer

	var ours []*build.Package
	for _, p := range pkg.Imports {
		if p == "C" {
			continue
		}
		innerPkg, err := build.Import(p, "", 0)
		if err != nil {
			log.Fatal(err)
		}
		if !isStdlib(innerPkg) && !isThirdParty(innerPkg) {
			ours = append(ours, innerPkg)
		}
	}

	numDeps[path] = len(ours)
	deps[path] = ours

	for _, innerPkg := range ours {
		analyze(innerPkg, layer+1)
	}
}

func isStdlib(p *build.Package) bool {
	if *isIncludeStdlib {
		return false
	}
	return p.Goroot
}

func isThirdParty(p *build.Package) bool {
	if *isIncludeLibs {
		return false
	}
	if strings.HasPrefix(p.ImportPath, rootPackage) {
		return false
	}
	for _, root := range thirdPartyRoots {
		if strings.HasPrefix(p.ImportPath, root) {
			return true
		}
	}
	return false
}

func layerDisplay() {
	layers := make([][]string, lowestLayer+1)
	for pkgName, lay := range layerPos {
		if layers[lay] == nil {
			layers[lay] = make([]string, 0, 1)
		}
		layers[lay] = append(layers[lay], pkgName)
	}
	for layer, pkgs := range layers {
		annotated := make([]string, 0, len(pkgs))
		for _, pkgName := range pkgs {
			annotated = append(annotated, fmt.Sprintf("%s %d", short(pkgName), numDeps[pkgName]))
		}
		sort.Strings(annotated)
		fmt.Printf("%d: %s\n", layer, strings.Join(annotated, ", "))
	}
}

func depsDisplay(pkgName string) {
	if len(deps[pkgName]) == 0 {
		fmt.Println("No internal dependencies")
	}
	for _, pkg := range deps[pkgName] {
		fmt.Println(" ", short(pkg.ImportPath))
	}
}

func deepDepsDisplay(pkgName string, depth int) {
	indent := strings.Repeat("| ", depth)
	fmt.Printf("%s%s\n", indent, short(pkgName))
	for _, pkg := range deps[pkgName] {
		//fmt.Printf("%s|-> %s\n", indent, pkg.ImportPath)
		deepDepsDisplay(pkg.ImportPath, depth+1)
	}
}

func bold(msg string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", msg)
}

func short(name string) string {
	if !*isShort {
		return name
	}
	if len(name) <= len(rootPackage) {
		return name
	}
	s := strings.Replace(name, rootPackage, "", 1)
	return strings.Trim(s, "/")
}
