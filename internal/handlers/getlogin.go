package handlers

import (
	"goth/internal/templates"
	"net/http"
)

type LoginHandler struct{}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := templates.Login("Login")
	err := templates.Layout(c, "Login").Render(r.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
