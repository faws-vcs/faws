package tracker

import (
	"net/http"

	"github.com/faws-vcs/faws/faws/repo/p2p/tracker/models"
)

func (s *Server) handle_ice_server_list(rw http.ResponseWriter, r *http.Request) {
	s.respond(rw, http.StatusOK, models.ICEServerList{
		ICEServers: s.config.ICEServers,
	})
}
