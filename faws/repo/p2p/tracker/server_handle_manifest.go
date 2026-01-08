package tracker

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker/models"
	"github.com/faws-vcs/faws/faws/validate"
)

func (s *Server) handle_manifest_upload(rw http.ResponseWriter, r *http.Request) {
	// validate the name of the topic
	topic_name := r.PathValue("topic")
	if !(len(topic_name) == 64 && validate.Hex(topic_name)) {
		s.respond(rw, http.StatusBadRequest, models.GenericResponse{Message: "invalid topic"})
		return
	}

	manifest_data, err := io.ReadAll(io.LimitReader(r.Body, 10e6))
	if err != nil {
		s.respond(rw, http.StatusRequestEntityTooLarge, models.GenericResponse{Message: "manifest data too large"})
		return
	}

	// parse manifest
	var manifest Manifest
	if err = DecodeManifest(manifest_data, &manifest); err != nil {
		s.respond(rw, http.StatusBadRequest, models.GenericResponse{})
		return
	}

	// verify
	if !identity.Verify(manifest.Publisher, &manifest.Signature, manifest.Info) {
		s.respond(rw, http.StatusBadRequest, models.GenericResponse{})
		return
	}
	if s.config.WhitelistPublishers {
		if !slices.Contains(s.config.PublisherWhitelist, manifest.Publisher) {
			console.Println("Publisher is unauthorized:", manifest.Publisher)
			s.respond(rw, http.StatusUnauthorized, models.GenericResponse{
				Message: "You are not authorized to publish to this tracker",
			})
			return
		}
	}

	// ensure bucket directory exists
	bucket := filepath.Join(s.directory, "manifests", topic_name[0:2])
	if _, err := os.Stat(bucket); err != nil {
		os.MkdirAll(bucket, fs.DefaultPrivateDirPerm)
	}

	s.guard_write.Lock()
	defer s.guard_write.Unlock()

	name := filepath.Join(bucket, topic_name[2:])
	// check if manifest already exists
	if _, err := os.Stat(name); err == nil {
		var existing_manifest Manifest
		existing_manifest_data, err := os.ReadFile(name)
		if err != nil {
			s.respond(rw, http.StatusBadRequest, models.GenericResponse{})
			return
		}
		// if the existing manifest matches
		if bytes.Equal(existing_manifest_data, manifest_data) {
			s.respond(rw, http.StatusOK, models.GenericResponse{})
			return
		}
		err = DecodeManifest(existing_manifest_data, &existing_manifest)
		if err != nil {
			s.respond(rw, http.StatusBadRequest, models.GenericResponse{})
			return
		}
		if existing_manifest.Time().After(manifest.Time()) {
			s.respond(rw, http.StatusConflict, models.GenericResponse{})
			return
		}
	}

	os.WriteFile(name, manifest_data, fs.DefaultPrivatePerm)
	s.respond(rw, http.StatusOK, models.GenericResponse{})
}

func (s *Server) handle_manifest_download(rw http.ResponseWriter, r *http.Request) {
	// validate the name of the topic
	topic_name := r.PathValue("topic")
	if !(len(topic_name) == 64 && validate.Hex(topic_name)) {
		s.respond(rw, http.StatusBadRequest, models.GenericResponse{})
		return
	}

	// ensure bucket directory exists
	name := filepath.Join(s.directory, "manifests", topic_name[0:2], topic_name[2:])
	// check if manifest already exists
	if _, err := os.Stat(name); err == nil {
		rw.Header().Set("Content-Type", mime_faws_manifest)
	}

	http.ServeFile(rw, r, name)
}
