package main

func (s *Server) routes() {
	s.router.HandleFunc("/verifycred", s.verifycredentials()).Methods("POST")
	s.router.HandleFunc("/registeruser", s.handleregisteruser()).Methods("POST")
	s.router.HandleFunc("/assigngroup", s.handleassigngroup()).Methods("GET")
}
