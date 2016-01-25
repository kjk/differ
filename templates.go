package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
)

var (
	tmplIndex     = "index.html"
	templateNames = []string{tmplIndex}
	templatePaths []string
	templates     *template.Template

	reloadTemplates = true
)

func loadTemplate(path string) ([]byte, error) {
	if hasZipResources() {
		path = normalizePath(path)
		d := resourcesFromZip[path]
		if d != nil {
			return d, nil
		}
		return nil, fmt.Errorf("file '%s' not in zip file", path)
	}
	return ioutil.ReadFile(path)
}

func parseTemplates(filenames ...string) (*template.Template, error) {
	var t *template.Template
	for _, filename := range filenames {
		b, err := loadTemplate(filename)
		if err != nil {
			return nil, err
		}
		s := string(b)
		name := filepath.Base(filename)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func getTemplates() *template.Template {
	if reloadTemplates || (nil == templates) {
		if 0 == len(templatePaths) {
			for _, name := range templateNames {
				templatePaths = append(templatePaths, filepath.Join("www", name))
			}
		}
		t, err := parseTemplates(templatePaths...)
		templates = template.Must(t, err)
	}
	return templates
}

func execTemplate(w http.ResponseWriter, templateName string, model interface{}) bool {
	var buf bytes.Buffer
	if err := getTemplates().ExecuteTemplate(&buf, templateName, model); err != nil {
		LogErrorf("Failed to execute template %q, error: %s", templateName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return false
	}

	// at this point we ignore error
	w.Header().Set("Content-Length", strconv.Itoa(len(buf.Bytes())))
	w.Write(buf.Bytes())
	return true
}
