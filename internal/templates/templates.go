package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bnuredini/telltime/internal/dbgen"
	"github.com/bnuredini/telltime/internal/services/activity"
	"github.com/bnuredini/telltime/ui"
)

type PageName string

const (
	baseTemplatePath     = "gohtml/base.gohtml"
	partialsTemplateGlob = "gohtml/partials/*.gohtml"
)

const (
	Page404      PageName = "404"
	Page500      PageName = "500"
	PageHome     PageName = "home"
	PageSettings PageName = "settings"
	PageActivity PageName = "activity"
	PageDocs     PageName = "docs"
)

type Data struct {
	WindowChangeEvents []dbgen.GetEventsRow
	CategoryStats      []*activity.CategoryStat
	ProgramStats       []*activity.ProgramStat
	CurrentDateLabel   string
	SelectedDateParam  string
	ScreenTimeSecs     int64
	ApplicationsUsed   int
	ContextSwitches    int
	LongestSessionSecs int64
	MostUsedProgram    string
	TopCategory        string
	CalendarData       *CalendarData
}

type Manager struct {
	PageCache        map[string]*template.Template
	DocCache         map[string]*template.Template
	PartialsTemplate *template.Template
}

func NewManager() (*Manager, error) {
	pageCache, err := newPageCache()
	if err != nil {
		return nil, err
	}

	docCache, err := newDocCache()
	if err != nil {
		return nil, err
	}

	partialTemplate, err := template.New("base").Funcs(tmplFuncs).ParseFS(ui.Files, partialsTemplateGlob)
	if err != nil {
		return nil, err
	}

	return &Manager{
		PageCache:        pageCache,
		DocCache:         docCache,
		PartialsTemplate: partialTemplate,
	}, nil
}

func NewData() *Data {
	return &Data{}
}

func newPageCache() (map[string]*template.Template, error) {
	return generateCacheFromGlob(
		"gohtml/pages/*.gohtml",
		func(filePath string) []string {
			return []string{baseTemplatePath, partialsTemplateGlob, filePath}
		},
	)
}

func newDocCache() (map[string]*template.Template, error) {
	return generateCacheFromGlob(
		"gohtml/docs/*.gohtml",
		func(filePath string) []string {
			return []string{baseTemplatePath, partialsTemplateGlob, "gohtml/pages/docs.gohtml", filePath}
		},
	)
}

// generateCacheFromGlob builds a cache of templates. Every key is derived from
// a template found in globPattern and every associated value is a template set
// that contains that template plus the files returned by buildTemplateBits.
func generateCacheFromGlob(
	globPattern string,
	buildTemplateBits func(filePath string) []string,
) (map[string]*template.Template, error) {
	filePaths, err := fs.Glob(ui.Files, globPattern)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*template.Template, len(filePaths))
	for _, filePath := range filePaths {
		bits := buildTemplateBits(filePath)
		tmpl, err := template.New("base").Funcs(tmplFuncs).ParseFS(ui.Files, bits...)
		if err != nil {
			return nil, err
		}

		result[cacheKeyFromPath(filePath)] = tmpl
	}

	return result, nil
}

func cacheKeyFromPath(filePath string) string {
	return strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
}

func RenderPage(
	manager *Manager,
	w http.ResponseWriter,
	pageName PageName,
	tmplData *Data,
) error {
	tmpl := manager.PageCache[string(pageName)]
	if tmpl == nil {
		tmpl = manager.PageCache[string(Page404)]
	}

	return writeTemplate(w, tmpl, "base", tmplData)
}

func RenderDoc(
	manager *Manager,
	w http.ResponseWriter,
	docName string,
	tmplData *Data,
) error {
	tmpl := manager.DocCache[string(docName)]
	if tmpl == nil {
		tmpl = manager.PageCache[string(Page404)]
	}

	return writeTemplate(w, tmpl, "base", tmplData)
}

func RenderPartial(
	manager *Manager,
	w http.ResponseWriter,
	partialName string,
	tmplData any,
) error {
	if manager.PartialsTemplate == nil {
		return fmt.Errorf("partial templates were not initialized")
	}

	if manager.PartialsTemplate.Lookup(partialName) == nil {
		return fmt.Errorf("partial template %q was not found in cache", partialName)
	}

	return writeTemplate(w, manager.PartialsTemplate, partialName, tmplData)
}

func writeTemplate(
	w http.ResponseWriter,
	tmpl *template.Template,
	templateName string,
	tmplData any,
) error {
	buf := new(bytes.Buffer)
	err := tmpl.ExecuteTemplate(buf, templateName, tmplData)
	if err != nil {
		return err
	}

	_, err = buf.WriteTo(w)

	return err
}
