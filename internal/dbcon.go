package internal

import(
"github.com/jackc/pgx/v5"
"log"
"context"
)

type PostgresDB struct{
	Conn *pgx.Conn
}

var pgdb *PostgresDB

func Setup(dburl string)error{
	pgdb = &PostgresDB{}
		
	log.Print("connecting postgres.")
	
	var err error
	pgdb.Conn,err = openConnection(dburl)
	if err != nil{
		return err
	}
	log.Print("connection successfull.")

	log.Print("ping?")
	if err := pgdb.Conn.Ping(context.Background()); err != nil{
		return err
	}
	log.Print("ping successfull")
	
	log.Print("migrating tables.")
	if err := pgdb.migrate();err != nil{
		return err
	}
	log.Print("migration successfull.")
	return nil
}


func (pg *PostgresDB) migrate()error{
	if err := pg.migrateUserTable(); err != nil{
		log.Print("couldnt migrate table user: ",err)
		return err
	}

	return nil
}

func openConnection(dburl string)(*pgx.Conn,error){
	conn,err := pgx.Connect(context.Background(),dburl)
	if err != nil{
		log.Print("couldnt connect to postgres: ",err)
		return nil,err
	}
	return conn,nil
}

