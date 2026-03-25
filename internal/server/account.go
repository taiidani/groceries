package server

import "net/http"

func (s *Server) accountHandler(w http.ResponseWriter, r *http.Request) {
	type accountBag struct {
		baseBag
	}

	bag := accountBag{baseBag: s.newBag(r.Context())}
	renderHtml(w, http.StatusOK, "account.gohtml", bag)
}
