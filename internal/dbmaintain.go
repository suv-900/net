package internal

import(
"github.com/jackc/pgx/v5"
"log"
)

//no locking tables as of ver 7
func (pg *PostgresDB) vaccumDB(ctx context.Context)error{
	sql := `VACCUM;`

	cmtag,err := pg.Conn.Exec(ctx,sql)
	if err != nil{
		log.Print("error while vaccuming: ",err)
		return err
	}
	log.Print("rows affected: ",cmtag.RowsAffected())

	return nil
}

func (pg *PostgresDB) vaccumUsers(ctx context.Context)error{
	sql := `VACCUM users;`

	cmtag,err := pg.Conn.Exec(ctx,sql)
	if err != nil{
		log.Print("error while vaccuming users: ",err)
		return err
	}
	log.Print("rows affected: ",cmtag.RowsAffected())

	return nil
}
