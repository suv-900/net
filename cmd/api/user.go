package api 

import(
"io"
"net/http"
)

func Register(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"register")	
}

func Login(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"login")	
}

func CheckExists(w http.ResponseWriter,r *http.Request){
	io.WriteString(w,"check")	
}
