package httphandler

import (
	"database/sql"
	"time"
	// "fmt"
	"log"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/bnuredini/telltime/internal/templates"
	"github.com/bnuredini/telltime/internal/services/activity"
)

type Handler struct {
	DB              *sql.DB
	TemplateManager *templates.TemplateManager
}

func New(db *sql.DB, templateManager *templates.TemplateManager) *Handler {
	return &Handler{
		DB:              db,
		TemplateManager: templateManager,
	}
}

func (h *Handler) HomeGet(w http.ResponseWriter, r *http.Request) {
	activity.Save(h.DB)

	start := time.Now().Add(-time.Hour)
	end := time.Now()
	programStats, err := activity.GetProgramStats(h.DB, start, end)
	if err != nil {

		return
	}

	tmplData := templates.NewData()
	tmplData.ProgramStats = programStats

	err = templates.RenderPage(h.TemplateManager, w, "home", tmplData)
	if err != nil {
		h.renderInternalServerError(w, r, err)
	}
}

func (h *Handler) ActivityGet(w http.ResponseWriter, r *http.Request) {
	activity.Save(h.DB)

	events, err := activity.GetEvents(h.DB)
	if err != nil {
		h.renderInternalServerError(w, r, err)
		return
	}

	tmplData := templates.NewData()
	tmplData.WindowChangeEvents = events

	err = templates.RenderPage(h.TemplateManager, w, "activity", tmplData)
	if err != nil {
		h.renderInternalServerError(w, r, err)
	}
}

func (h *Handler) renderInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s\n%s\n", err.Error(), debug.Stack())

	w.WriteHeader(http.StatusInternalServerError)

	err = templates.RenderPage(h.TemplateManager, w, "500", templates.NewData())
	if err != nil {
		slog.Error("rendering failed", "err", err)
	}
}
