package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/bnuredini/telltime/ui"
	"github.com/bnuredini/telltime/internal/services/activity"
)

type TemplateManager struct {
	Cache map[string]*template.Template
}

func NewManager() (*TemplateManager, error) {
	result := &TemplateManager{}

	cache, err := newCache()
	if err != nil {
		return nil, err
	}

	result.Cache = cache

	return result, nil
}

type templateData struct {
	WindowChangeEvents []*activity.WindowChangeEvent
	CategoryStats []*activity.CategoryStat
	ProgramStats []*activity.ProgramStat
}

func NewData() *templateData {
	return &templateData{}
}

var tmplFuncs = template.FuncMap{
	"now": time.Now,
}

func newCache() (map[string]*template.Template, error) {
	result := make(map[string]*template.Template)

	pages, err := fs.Glob(ui.Files, "html/pages/*.gohtml")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		pageFileName := filepath.Base(page)

		bits := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			page,
		}

		tmpl, err := template.New(pageFileName).Funcs(tmplFuncs).ParseFS(ui.Files, bits...)
		if err != nil {
			return nil, err
		}

		result[pageFileName] = tmpl
	}

	return result, nil
}

func RenderPage(
	manager *TemplateManager,
	w http.ResponseWriter,
	pageName string,
	tmplData *templateData,
) error {

	pageKey := fmt.Sprintf("%s.gohtml", pageName)
	tmpl := manager.Cache[pageKey]

	if tmpl == nil {
		slog.Info("couldn't get tempalte from cache; now generating it...")

		bits := []string{
			"html/base.gohtml",
			"html/partials/*.gohtml",
			fmt.Sprintf("html/pages/%s.gohtml", pageName),
		}

		tmplName := fmt.Sprintf("%s.gohtml", pageName)
		var err error
		tmpl, err = template.New(tmplName).Funcs(tmplFuncs).ParseFS(ui.Files, bits...)
		if err != nil {
			return err
		}
	}

	buf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(buf, "base", tmplData)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)

	return err
}
