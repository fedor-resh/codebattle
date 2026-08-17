package httpapi

import (
	"errors"
	"net/http"

	"codebattle.local/codebattle/internal/practice"
	"codebattle.local/codebattle/internal/submissions"
)

type practiceHandlers struct {
	accounts    accountHandlers
	practice    *practice.Service
	submissions *submissions.Repository
}

type startPracticeRequest struct {
	Slug string `json:"slug"`
}

type updatePracticeCodeRequest struct {
	SourceCode string `json:"source_code"`
	Revision   int64  `json:"revision"`
}

func (h practiceHandlers) problems(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	items, err := h.practice.Problems(r.Context(), user.ID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	if items == nil {
		items = []practice.ProblemSummary{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"problems": items})
}

func (h practiceHandlers) startSession(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	var input startPracticeRequest
	if err := decodeJSON(r, &input); err != nil || input.Slug == "" {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Укажите slug задачи", nil)
		return
	}
	session, err := h.practice.StartSession(r.Context(), user.ID, input.Slug)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"session": session})
}

func (h practiceHandlers) session(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	session, err := h.practice.Session(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session": session})
}

func (h practiceHandlers) updateCode(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	var input updatePracticeCodeRequest
	if err := decodeJSON(r, &input); err != nil || input.Revision < 0 {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Передайте source_code и revision", nil)
		return
	}
	if err := h.practice.UpdateCode(
		r.Context(), user.ID, r.PathValue("id"), input.SourceCode, input.Revision,
	); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h practiceHandlers) createSubmission(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	var input createSubmissionRequest
	if err := decodeJSON(r, &input); err != nil || input.SourceCode == "" {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Передайте source_code", nil)
		return
	}
	submission, err := h.submissions.CreatePractice(r.Context(), user.ID, r.PathValue("id"), input.SourceCode)
	if err != nil {
		submissionHandlers{}.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"submission": submission})
}

func (h practiceHandlers) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, practice.ErrProblemNotFound):
		writeError(w, r, http.StatusNotFound, "PROBLEM_NOT_FOUND", "Задача не найдена", nil)
	case errors.Is(err, practice.ErrSessionNotFound):
		writeError(w, r, http.StatusNotFound, "SESSION_NOT_FOUND", "Сессия тренировки не найдена", nil)
	case errors.Is(err, practice.ErrStaleRevision):
		writeError(w, r, http.StatusConflict, "STALE_REVISION", "Получена устаревшая версия кода", nil)
	case errors.Is(err, practice.ErrSourceTooLarge):
		writeError(w, r, http.StatusRequestEntityTooLarge, "SOURCE_TOO_LARGE", "Максимальный размер исходника — 64 КБ", nil)
	default:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
	}
}
