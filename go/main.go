package main

import (
	"errors"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/swaggo/swag"
)

var (
	tmpl = template.Must(template.New("main").Parse(mainTextTemplate))
)

var (
	mainTextTemplate = `package main
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io
	
// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
func main() {}	
`
)

func main() {
	packit.Run(detectFn, buildFn)
}

func detectFn(ctx packit.DetectContext) (packit.DetectResult, error) {
	pattern := "*.go"

	errFileMatch := errors.New("File matched")
	err := filepath.Walk(ctx.WorkingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		match, err := filepath.Match(pattern, info.Name())
		if err != nil {
			return err
		}
		if match {
			return errFileMatch
		}
		return nil
	})

	if err == errFileMatch {
		return packit.DetectResult{}, nil
	}

	if err != nil {
		return packit.DetectResult{}, errors.New("no go file found.")
	}

	return packit.DetectResult{}, err
}

func buildFn(ctx packit.BuildContext) (packit.BuildResult, error) {
	err := createMainGoFile(ctx.WorkingDir)
	if err != nil {
		return packit.BuildResult{}, err
	}

	p := swag.New()
	err = p.ParseAPI(ctx.WorkingDir, "main.go", 2)
	if err != nil {
		return packit.BuildResult{}, err
	}

	body, err := p.GetSwagger().MarshalJSON()
	if err != nil {
		return packit.BuildResult{}, err
	}

	fd, err := os.Create(filepath.Join(ctx.Layers.Path, "swag.json"))
	if err != nil {
		return packit.BuildResult{}, err
	}
	defer fd.Close()

	_, err = fd.Write(body)
	if err != nil {
		return packit.BuildResult{}, err
	}
	err = fd.Sync()
	if err != nil {
		return packit.BuildResult{}, err
	}
	return packit.BuildResult{}, nil
}

func createMainGoFile(path string) error {
	f, err := os.Create(filepath.Join(path, "main.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, nil); err != nil {
		return fmt.Errorf("executing template: %v", err)
	}
	return nil
}
