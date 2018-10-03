package api

import "github.com/gorilla/mux"

func (a *WebAPI) initEndpoints(r *mux.Router, apiServer *WebAPI) {
	r.HandleFunc("/cluster/join", apiServer.ClusterJoin).Methods("POST")
	r.HandleFunc("/cluster/leave", apiServer.ClusterLeave).Methods("GET")
	r.HandleFunc("/add/{key}", apiServer.AddObject).Methods("POST")
	r.HandleFunc("/del/{key}", apiServer.RemObject).Methods("DELETE")
	r.HandleFunc("/get/{key}", apiServer.LoadObject).Methods("GET")
}
