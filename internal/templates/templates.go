package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/bnuredini/telltime/ui"
	"github.com/bnuredini/telltime/internal/dbgen"
	"github.com/bnuredini/telltime/internal/services/activity"
)

type PageName string

const (
	Page500      PageName = "500"
	PageHome     PageName = "home"
	PageActivity PageName = "activity"
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
	WindowChangeEvents []dbgen.GetEventsRow
	CategoryStats      []*activity.CategoryStat
	ProgramStats       []*activity.ProgramStat
}

func NewData() *templateData {
	return &templateData{}
}

var tmplFuncs = template.FuncMap{
	"now":        time.Now,
	"formatSecs": formatSecs,
}

func formatSecs(secs int64) string {
	formattedHours := secs / 3600
	formattedMins := (secs % 3600) / 60
	formattedSecs := secs % 60

	var b strings.Builder

	if formattedHours > 0 {
		fmt.Fprintf(&b, "%dh", formattedHours)
	}
	if formattedMins > 0 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%dm", formattedMins)
	}
	if formattedSecs > 0 || b.Len() == 0 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, "%ds", formattedSecs)
	}

	return b.String()
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
	pageName PageName,
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
