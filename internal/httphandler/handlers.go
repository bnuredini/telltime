package httphandler

import (
	"context"
	"time"
	"database/sql"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/bnuredini/telltime/internal/dbgen"
	"github.com/bnuredini/telltime/internal/services/activity"
	"github.com/bnuredini/telltime/internal/templates"
)

type Handler struct {
	DB              *sql.DB
	Queries         *dbgen.Queries
	TemplateManager *templates.Manager
}

func New(db *sql.DB, queries *dbgen.Queries, templateManager *templates.Manager) *Handler {
	return &Handler{
		DB:              db,
		Queries:         queries,
		TemplateManager: templateManager,
	}
}

func (h *Handler) HomeGet(w http.ResponseWriter, r *http.Request) {
	currDate := parseDate(
		r.URL.Query().Get("curr-date"),
		r.URL.Query().Get("time-zone"),
		time.Now(),
	)

	if err := activity.Save(h.DB); err != nil {
		slog.Error("serving home: failed to save activty data", "err", err)
	}

	start, end := activity.GetIntervalFromStartOfDay()
	programStats, err := activity.GetProgramStats(context.Background(), h.Queries, start, end)
	if err != nil {
		h.renderInternalServerError(w, r, err)
		return
	}

	tmplData := templates.NewData()
	tmplData.ProgramStats = programStats
	tmplData.CalendarData = templates.NewCalendarData(currDate)

	err = templates.RenderPage(h.TemplateManager, w, templates.PageHome, tmplData)
	if err != nil {
		h.renderInternalServerError(w, r, err)
	}
}

func (h *Handler) ActivityGet(w http.ResponseWriter, r *http.Request) {
	if err := activity.Save(h.DB); err != nil {
		slog.Error("serving page for activities: failed to save activty data", "err", err)
	}
	// TODO: Format these values for the frontend. The NullString values should
	// be sent to the templates. Also consider adding template function instead.
	events, err := h.Queries.GetEvents(context.Background())
	if err != nil {
		h.renderInternalServerError(w, r, err)
		return
	}

	tmplData := templates.NewData()
	tmplData.WindowChangeEvents = events

	err = templates.RenderPage(h.TemplateManager, w, templates.PageActivity, tmplData)
	if err != nil {
		h.renderInternalServerError(w, r, err)
	}
}

func (h *Handler) CalendarSelectGet(w http.ResponseWriter, r *http.Request) {
	currDate := parseDate(
		r.URL.Query().Get("curr-date"),
		r.URL.Query().Get("time-zone"),
		time.Now(),
	)

	tmplData := templates.NewData()
	tmplData.CalendarData = templates.NewCalendarData(currDate)

	err := templates.RenderPartial(h.TemplateManager, w, "calendar", tmplData)
	if err != nil {
		h.renderInternalServerError(w, r, err)
	}
}

func (h *Handler) renderInternalServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal server error", "err", err.Error(), "stack", debug.Stack())
	w.WriteHeader(http.StatusInternalServerError)
	// TODO: Pass the Request struct value to the template.
	err = templates.RenderPage(h.TemplateManager, w, templates.Page500, templates.NewData())
	if err != nil {
		slog.Error("rendering failed", "err", err)
	}
}
