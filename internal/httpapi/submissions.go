package httpapi

import (
	"errors"
	"net/http"

	"codebattle.local/codebattle/internal/submissions"
)

type submissionHandlers struct {
	accounts   accountHandlers
	repository *submissions.Repository
}

type createSubmissionRequest struct {
	SourceCode string `json:"source_code"`
}

func (h submissionHandlers) create(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	var input createSubmissionRequest
	if err := decodeJSON(r, &input); err != nil || input.SourceCode == "" {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Передайте source_code", nil)
		return
	}
	submission, err := h.repository.Create(r.Context(), user.ID, r.PathValue("id"), input.SourceCode)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"submission": submission})
}

func (h submissionHandlers) get(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	submission, err := h.repository.Get(r.Context(), r.PathValue("id"), user.ID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"submission": submission})
}

func (h submissionHandlers) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, submissions.ErrInvalidSource):
		writeError(w, r, http.StatusUnprocessableEntity, "INVALID_SOURCE", err.Error(), nil)
	case errors.Is(err, submissions.ErrSourceTooLarge):
		writeError(w, r, http.StatusRequestEntityTooLarge, "SOURCE_TOO_LARGE", "Максимальный размер исходника — 64 КБ", nil)
	case errors.Is(err, submissions.ErrMatchNotFound):
		writeError(w, r, http.StatusConflict, "MATCH_NOT_ACTIVE", "Матч не найден или уже завершён", nil)
	case errors.Is(err, submissions.ErrSessionNotFound):
		writeError(w, r, http.StatusNotFound, "SESSION_NOT_FOUND", "Сессия тренировки не найдена", nil)
	case errors.Is(err, submissions.ErrRateLimited):
		writeError(w, r, http.StatusTooManyRequests, "SUBMISSION_RATE_LIMIT", "Подождите 2 секунды перед следующей отправкой", nil)
	case errors.Is(err, submissions.ErrTooManyPending):
		writeError(w, r, http.StatusConflict, "TOO_MANY_PENDING", "Допускается не более трёх незавершённых отправок", nil)
	case errors.Is(err, submissions.ErrSubmissionNotFound):
		writeError(w, r, http.StatusNotFound, "SUBMISSION_NOT_FOUND", "Отправка не найдена", nil)
	default:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
	}
}
