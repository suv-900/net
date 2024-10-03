package api

import(
"log"
"github.com/suv-900/net/internal"
)

type App struct{

}

var app *App

func Start(){
	app = &App{}
}

func (a *App) setupDB()error{

	dburl,exists := os.LookupEnv("dburl")
	if !exists{
		return errors.New("env var(dburl) doesnt exists.")
	}
	
	err := internal.Setup(dburl)
	if err != nil{
		return err
	}
}
