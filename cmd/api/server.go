package api 

import(
"log"
"net/http"
"github.com/gorilla/mux"
)

func SetupServer(){
	rmux := mux.NewRouter()
	
	log.Print("setting up router")
	setup_router(rmux)

	srv := &http.Server{
		Addr: "127.0.0.1:8000",
		Handler:rmux,	
	}
	log.Print("starting web server")

	log.Fatal(srv.ListenAndServe())
}
func OK(){}
func setup_router(r *mux.Router){
	r.HandleFunc("/register",Register)
	r.HandleFunc("/login",Login)
	r.HandleFunc("/exists",CheckExists)
}

