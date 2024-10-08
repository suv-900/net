package internal

import(
"github.com/jackc/pgx/v5"
"log"
"context"
)

func CreatePostgresConn(ctx context.Context,dburl string)(*pgx.Conn,error){
	log.Print("connecting postgres.")
	conn,err := openConnection(ctx,dburl)
	if err != nil{
		log.Print("connection unsuccessfull: ",err)
		return nil,err
	}
	log.Print("connection successfull.")

	log.Print("ping?")
	if err := conn.Ping(context.Background()); err != nil{
		log.Print("ping unsuccessfull: ",err)
		return nil,err
	}
	log.Print("ping successfull")
	
	return conn,nil 
}

func openConnection(ctx context.Context,dburl string)(*pgx.Conn,error){
	conn,err := pgx.Connect(ctx,dburl)
	if err != nil{
		log.Print("couldnt connect to postgres: ",err)
		return nil,err
	}
	return conn,nil
}
