package httpapi

import (
	"errors"
	"net/http"

	"codebattle.local/codebattle/internal/duels"
)

type duelHandlers struct {
	accounts accountHandlers
	service  *duels.Service
}

type createInvitationRequest struct {
	ReceiverID string `json:"receiver_id"`
}

type updateCodeRequest struct {
	SourceCode   string `json:"source_code"`
	Revision     int64  `json:"revision"`
	CursorLine   int    `json:"cursor_line"`
	CursorColumn int    `json:"cursor_column"`
}

func (h duelHandlers) createInvitation(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	var input createInvitationRequest
	if err := decodeJSON(r, &input); err != nil || input.ReceiverID == "" {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Укажите receiver_id", nil)
		return
	}
	invitation, err := h.service.CreateInvitation(r.Context(), user.ID, input.ReceiverID)
	if err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"invitation": invitation})
}

func (h duelHandlers) invitationState(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	state, err := h.service.State(r.Context(), user.ID)
	if err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h duelHandlers) acceptInvitation(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	match, err := h.service.AcceptInvitation(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"match": match})
}

func (h duelHandlers) declineInvitation(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	if err := h.service.DeclineInvitation(r.Context(), user.ID, r.PathValue("id")); err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h duelHandlers) match(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	match, err := h.service.Match(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"match": match})
}

func (h duelHandlers) leaveMatch(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	if err := h.service.LeaveMatch(r.Context(), user.ID, r.PathValue("id")); err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h duelHandlers) updateCode(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	var input updateCodeRequest
	if err := decodeJSON(r, &input); err != nil || input.Revision < 0 {
		writeError(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Передайте source_code и revision", nil)
		return
	}
	// Keep the endpoint compatible with clients deployed before cursor sharing.
	if input.CursorLine == 0 {
		input.CursorLine = 1
	}
	if input.CursorColumn == 0 {
		input.CursorColumn = 1
	}
	if err := h.service.UpdateCode(
		r.Context(), user.ID, r.PathValue("id"), input.SourceCode, input.Revision,
		input.CursorLine, input.CursorColumn,
	); err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h duelHandlers) ready(w http.ResponseWriter, r *http.Request) {
	user, ok := h.accounts.authenticate(w, r)
	if !ok {
		return
	}
	match, err := h.service.Ready(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		h.writeDuelError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"match": match})
}

func (h duelHandlers) writeDuelError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, duels.ErrSelfInvitation):
		writeError(w, r, http.StatusUnprocessableEntity, "SELF_INVITATION", "Нельзя пригласить самого себя", nil)
	case errors.Is(err, duels.ErrUserNotFound):
		writeError(w, r, http.StatusNotFound, "USER_NOT_FOUND", "Пользователь не найден", nil)
	case errors.Is(err, duels.ErrUserUnavailable):
		writeError(w, r, http.StatusConflict, "USER_UNAVAILABLE", "Пользователь offline или уже находится в матче", nil)
	case errors.Is(err, duels.ErrInvitationBusy):
		writeError(w, r, http.StatusConflict, "INVITATION_CONFLICT", "У одного из игроков уже есть приглашение", nil)
	case errors.Is(err, duels.ErrInvitationGone):
		writeError(w, r, http.StatusConflict, "INVITATION_GONE", "Приглашение истекло или уже обработано", nil)
	case errors.Is(err, duels.ErrForbidden):
		writeError(w, r, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав для этого действия", nil)
	case errors.Is(err, duels.ErrMatchNotFound):
		writeError(w, r, http.StatusNotFound, "MATCH_NOT_FOUND", "Матч не найден", nil)
	case errors.Is(err, duels.ErrProblemsMissing):
		writeError(w, r, http.StatusServiceUnavailable, "PROBLEMS_UNAVAILABLE", "Каталог задач ещё не загружен", nil)
	case errors.Is(err, duels.ErrRoundNotActive):
		writeError(w, r, http.StatusConflict, "ROUND_NOT_ACTIVE", "Раунд уже завершён", nil)
	case errors.Is(err, duels.ErrStaleRevision):
		writeError(w, r, http.StatusConflict, "STALE_REVISION", "Получена устаревшая версия кода", nil)
	case errors.Is(err, duels.ErrSourceTooLarge):
		writeError(w, r, http.StatusRequestEntityTooLarge, "SOURCE_TOO_LARGE", "Максимальный размер исходника — 64 КБ", nil)
	case errors.Is(err, duels.ErrInvalidCursor):
		writeError(w, r, http.StatusBadRequest, "INVALID_CURSOR", "Некорректная позиция курсора", nil)
	default:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
	}
}
