package api

import "github.com/gorilla/mux"

func (a *WebAPI) initEndpoints(r *mux.Router, apiServer *WebAPI) {
	r.HandleFunc("/cluster/join", apiServer.ClusterJoin).Methods("GET")
	r.HandleFunc("/cluster/leave", apiServer.ClusterLeave).Methods("GET")
	r.HandleFunc("/add/{key}", apiServer.AddObject).Methods("POST")
	r.HandleFunc("/get/{key}", apiServer.RemObject).Methods("GET")
	r.HandleFunc("/del/{key}", apiServer.LoadObject).Methods("DELETE")
}
