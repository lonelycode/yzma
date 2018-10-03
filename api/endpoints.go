package api

import "github.com/gorilla/mux"

func (a *WebAPI) initEndpoints(r *mux.Router, apiServer *WebAPI) {
	r.HandleFunc("/cluster/join", apiServer.ClusterJoin).Methods("POST")
	r.HandleFunc("/cluster/leave", apiServer.ClusterLeave).Methods("POST")
	r.HandleFunc("/keys/{key}", apiServer.AddObject).Methods("POST")
	r.HandleFunc("/keys/{key}", apiServer.RemObject).Methods("DELETE")
	r.HandleFunc("/keys/{key}", apiServer.LoadObject).Methods("GET")
}
