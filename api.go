package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	store Store
}

func NewAPIServer(addr string, store Store) *APIServer {
	return &APIServer{addr: addr, store: store}
}

func (s*APIServer) Serve() {
	 router := mux.NewRouter()
	 subrouter := router.PathPrefix("/api/v1").Subrouter()


	usersService := NewUserService(s.store)
	usersService.RegisterRoutes(subrouter)

	// hi̇zmetleri̇mi̇zi̇n kayit altina alinmasi
	tasksService := NewTasksService(s.store)
	tasksService.RegisterRoutes(subrouter)


	log.Println("API sunucusunu şu adresten başlatıyorum:", s.addr)

	log.Fatal(http.ListenAndServe(s.addr, subrouter))
}