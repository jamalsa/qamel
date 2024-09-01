package generator

import (
	"bufio"
	"fmt"
	"os"
	fp "path/filepath"
	"regexp"
	"strings"

	"github.com/jamalsa/qamel/internal/config"
)

var rxQmlImport = regexp.MustCompile(`^import\s+(Qt\S+)\s(\d)+.*$`)

// copyQmlDependencies copies QML dependencies to output dir
func copyQmlDependencies(qtQmlDir string, profile config.Profile, projectDir, outputDir string) error {
	// Get list of dir to check
	var dirToCheck []string
	projectResDir := fp.Join(projectDir, "res")

	if dirExists(projectResDir) {
		dirToCheck = append(dirToCheck, projectResDir)
	}

	// Get list of QML dependencies
	mapDependencies := map[string]struct{}{}
	dirAlreadyChecked := map[string]struct{}{}

	for {
		// Keep checking until there are no directory left
		if len(dirToCheck) == 0 {
			break
		}

		// If this dir has been checked before, skip
		dirPath := dirToCheck[0]
		if _, checked := dirAlreadyChecked[dirPath]; checked {
			dirToCheck = dirToCheck[1:]
			continue
		}

		// Fetch qml dependencies from the directory
		qmlDeps, err := getQmlDependenciesFromDir(qtQmlDir, dirPath)
		if err != nil {
			return err
		}

		for _, qmlDep := range qmlDeps {
			mapDependencies[qmlDep] = struct{}{}
			dirToCheck = append(dirToCheck, qmlDep)
		}

		// Save this dir as been already checked, then move to next
		dirAlreadyChecked[dirPath] = struct{}{}
		dirToCheck = dirToCheck[1:]
	}

	// Copy all dependency libs to output dir
	dstQmlDir := outputDir
	if profile.OS != "windows" {
		dstQmlDir = fp.Join(dstQmlDir, "qml")
	}

	for qmlPath := range mapDependencies {
		parentExists := false
		qmlPathParts := fp.SplitList(qmlPath)
		for i := 0; i <= len(qmlPathParts)-1; i++ {
			parentPath := fp.Join(qmlPathParts[:i]...)
			if _, exist := mapDependencies[parentPath]; exist {
				parentExists = true
				break
			}
		}

		if parentExists {
			continue
		}

		dirName, err := fp.Rel(qtQmlDir, qmlPath)
		if err != nil {
			return err
		}

		err = copyDir(qmlPath, fp.Join(dstQmlDir, dirName), func(path string, info os.FileInfo) bool {
			fileExt := fp.Ext(info.Name())
			if fileExt == ".qmlc" || fileExt == ".jsc" {
				return true
			}

			if strings.HasSuffix(info.Name(), "d.dll") {
				tmpPath := strings.TrimSuffix(path, "d.dll")
				tmpPath += ".dll"

				if fileExists(tmpPath) {
					return true
				}
			}

			return false
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// getQmlDependenciesFromDir gets all QML imports from *.qml files
// inside specified directory
func getQmlDependenciesFromDir(qtQmlDir string, srcDir string) ([]string, error) {
	// Make sure dir existst
	if !dirExists(srcDir) {
		return nil, fmt.Errorf("directory %s doesn't exist", srcDir)
	}

	// Fetch each QML file inside the specified dir
	var qmlFiles []string
	fp.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && fp.Ext(info.Name()) == ".qml" {
			qmlFiles = append(qmlFiles, path)
		}
		return nil
	})

	// Get QML dependencies from each file
	mapDependencies := map[string]struct{}{}
	for _, qmlFile := range qmlFiles {
		deps, err := getQmlDependenciesFromFile(qtQmlDir, qmlFile)
		if err != nil {
			return nil, err
		}

		for _, dep := range deps {
			mapDependencies[dep] = struct{}{}
		}
	}

	// Convert map to arrays
	var results []string
	for depName := range mapDependencies {
		results = append(results, depName)
	}

	return results, nil
}

// getQmlDependenciesFromFile gets all QML imports from *.qml files
// from the specified file
func getQmlDependenciesFromFile(qtQmlDir string, qmlFilePath string) ([]string, error) {
	// Open QML file
	qmlFile, err := os.Open(qmlFilePath)
	if err != nil {
		return nil, err
	}
	defer qmlFile.Close()

	// Read each line from the file
	var results []string
	scanner := bufio.NewScanner(qmlFile)
	for scanner.Scan() {
		// Use regex to find import line
		line := scanner.Text()
		line = strings.TrimSpace(line)
		matches := rxQmlImport.FindStringSubmatch(line)

		// If regex doesn't match, skip
		if len(matches) != 3 {
			continue
		}

		// Find possible directory for that QML import from Qt's QML dir
		// For example, "import QtQuick.Controls 2.11" might means the QML files located in :
		// - qtQmlDir/QtQuick/Controls.2
		// - qtQmlDir/QtQuick.2/Controls
		// Or the version number might be not used afterall
		// - qtQmlDir/QtQuick/Controls
		name := matches[1]
		nameParts := strings.Split(name, ".")

		possibleDir := ""
		version := "." + matches[2]
		for i := len(nameParts) - 1; i >= 0; i-- {
			tmp := make([]string, len(nameParts))
			copy(tmp, nameParts)

			tmp[i] += version
			tmpDir := fp.Join(tmp...)
			tmpDir = fp.Join(qtQmlDir, tmpDir)
			if dirExists(tmpDir) {
				possibleDir = tmpDir
				break
			}
		}

		if possibleDir != "" {
			results = append(results, possibleDir)
		} else {
			possibleDir := fp.Join(nameParts...)
			possibleDir = fp.Join(qtQmlDir, possibleDir)
			if dirExists(possibleDir) {
				results = append(results, possibleDir)
			}
		}
	}

	return results, nil
}
