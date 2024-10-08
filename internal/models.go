package internal

import(
"context"
"os"
"log"
)
type Models struct{
	User interface{
		
	}
}

const(
	ErrInternalServerError = errors.New("internal server error.")
)

func CreateModel()(*Models,error){
	
	dburl,exists := os.LookupEnv("dburl")
	if !exists{
		log.Print("env var dburl doesnt exists")
		return nil,ErrInternalServerError
	}
	
	conn,err := CreatePGConn(context.Background(),dburl)
	if err != nil{
		return nil,err
	}
	
	models := Models{
		User: UserRepo{conn:conn}
	}
	
	if  err := models.User.Migrate(context.Background()); err != nil{
		return nil,ErrInternalServerError
	}

	return *models,nil
}

