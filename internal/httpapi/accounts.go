package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"codebattle.local/codebattle/internal/accounts"
)

const sessionCookieName = "codebattle_session"

type accountHandlers struct {
	service       *accounts.Service
	secureCookies bool
}

type credentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h accountHandlers) register(w http.ResponseWriter, r *http.Request) {
	var input credentialsRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Проверьте формат запроса", nil)
		return
	}
	session, err := h.service.Register(r.Context(), input.Username, input.Password)
	if err != nil {
		h.writeAccountError(w, r, err)
		return
	}
	h.setSessionCookie(w, session)
	writeJSON(w, http.StatusCreated, map[string]any{"user": session.User})
}

func (h accountHandlers) login(w http.ResponseWriter, r *http.Request) {
	var input credentialsRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "Проверьте формат запроса", nil)
		return
	}
	session, err := h.service.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		h.writeAccountError(w, r, err)
		return
	}
	h.setSessionCookie(w, session)
	writeJSON(w, http.StatusOK, map[string]any{"user": session.User})
}

func (h accountHandlers) logout(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie(sessionCookieName)
	if cookie != nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось завершить сессию", nil)
			return
		}
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h accountHandlers) me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticate(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h accountHandlers) heartbeat(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authenticate(w, r)
	if !ok {
		return
	}
	if err := h.service.Heartbeat(r.Context(), user.ID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось обновить статус", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h accountHandlers) users(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.authenticate(w, r); !ok {
		return
	}
	limit := 50
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			writeError(w, r, http.StatusBadRequest, "INVALID_LIMIT", "limit должен быть числом", nil)
			return
		}
		limit = parsed
	}
	page, err := h.service.ListUsers(
		r.Context(),
		r.URL.Query().Get("q"),
		r.URL.Query().Get("cursor"),
		limit,
	)
	if err != nil {
		if errors.Is(err, accounts.ErrInvalidCursor) {
			writeError(w, r, http.StatusBadRequest, "INVALID_CURSOR", "Некорректный cursor", nil)
			return
		}
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось загрузить пользователей", nil)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func (h accountHandlers) authenticate(w http.ResponseWriter, r *http.Request) (accounts.User, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется вход", nil)
		return accounts.User{}, false
	}
	user, err := h.service.Authenticate(r.Context(), cookie.Value)
	if err != nil {
		if !errors.Is(err, accounts.ErrUnauthorized) {
			writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Не удалось проверить сессию", nil)
			return accounts.User{}, false
		}
		h.clearSessionCookie(w)
		writeError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Сессия истекла", nil)
		return accounts.User{}, false
	}
	return user, true
}

func (h accountHandlers) writeAccountError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, accounts.ErrInvalidUsername):
		writeError(w, r, http.StatusUnprocessableEntity, "INVALID_USERNAME", "Username: 3–24 латинских символа, цифры или _", nil)
	case errors.Is(err, accounts.ErrInvalidPassword):
		writeError(w, r, http.StatusUnprocessableEntity, "INVALID_PASSWORD", "Пароль должен содержать от 6 до 128 символов", nil)
	case errors.Is(err, accounts.ErrUsernameTaken):
		writeError(w, r, http.StatusConflict, "USERNAME_TAKEN", "Этот username уже занят", nil)
	case errors.Is(err, accounts.ErrInvalidCredentials):
		writeError(w, r, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Неверный username или пароль", nil)
	default:
		writeError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера", nil)
	}
}

func (h accountHandlers) setSessionCookie(w http.ResponseWriter, session accounts.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h accountHandlers) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 64*1024))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
