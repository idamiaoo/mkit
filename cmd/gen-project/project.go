package genproject

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gobuffalo/packr/v2"
	generatego "github.com/gogo/protobuf/protoc-gen-gogo/generator"
)

type project struct {
	Name         string
	ModPrefix    string
	HttpPort     int
	GrpcPort     int
	AbsolutePath string
}

// funcMap contains a series of utility functions to be passed into
// templates and used within those templates.
var funcMap = template.FuncMap{
	"ToLower": strings.ToLower,
	"GoName":  generatego.CamelCase,
}

func (p *project) create() (err error) {
	box := packr.New("templates", "./templates")
	if err = os.MkdirAll(p.AbsolutePath, 0755); err != nil {
		return
	}
	for _, name := range box.List() {
		tmpl, _ := box.FindString(name)
		i := strings.LastIndex(name, string(os.PathSeparator))
		if i > 0 {
			dir := name[:i]
			if err = os.MkdirAll(filepath.Join(p.AbsolutePath, dir), 0755); err != nil {
				return
			}
		}
		if strings.HasSuffix(name, ".tmpl") {
			name = strings.TrimSuffix(name, ".tmpl")
		}
		if err = p.write(filepath.Join(p.AbsolutePath, name), tmpl); err != nil {
			return
		}
	}

	if err = p.generate("./..."); err != nil {
		return
	}
	if err = p.generate("./internal/dao/wire.go"); err != nil {
		return
	}
	return
}

func (p *project) generate(path string) error {
	cmd := exec.Command("go", "generate", "-x", path)
	cmd.Dir = p.AbsolutePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *project) write(path, tpl string) (err error) {
	data, err := p.parse(tpl)
	if err != nil {
		return
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (p *project) parse(s string) ([]byte, error) {
	t, err := template.New("").Funcs(funcMap).Parse(s)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err = t.Execute(&buf, p); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
