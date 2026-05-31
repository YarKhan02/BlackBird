package handler

import "net/http"

// Apps hit this to get the public key and verify JWTs themselves
// Zero dependency on your auth server at request time
func (h *AuthHandler) JWKS(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.tokenSvc.JWKS())
}
