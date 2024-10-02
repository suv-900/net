package internal

import(
"github.com/jackc/pgx/v5"
"log"
)

type PGDB struct{
	conn *pgx.Conn
}

var pgdb *PGDB

func setup()error{
	pgdb = &PGDB{}
	
	log.Print("connecting postgres.")
	if pgdb.conn,err := openConnection();err != nil{
		return err
	}
	log.Print("connection successfull.")

	log.Print("sending ping.")
	if err := pgdb.conn.Ping(context.Background()); err != nil{
		return err
	}
	log.Print("ping successfull")
	
	log.Print("migrating tables.")
	if err := pgdb.migrate();err != nil{
		return err
	}

	return nil
}


func (pg *PostgresDB) migrate()error{
	if err := pg.migrateUserTable(); err != nil{
		log.Print("couldnt migrate table user: ",err)
		return err
	}

	return nil
}

func openConnection()(*pgx.Conn,error){
	dburl := "postgres://core:123@localhost:5432/net"
	conn,err := pgx.Connect(context.Background(),dburl)
	if err != nil{
		log.Print("couldnt connect to postgres: ",err)
		return err
	}
	return conn
}

