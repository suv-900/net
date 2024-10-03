package internal

import(
"time"
"context"
"github.com/jackc/pgx/v5"
)

type User struct{
	Id uint 
	Name string
	Email string
	EmailVerified bool
	Password string
	Role string
	Bio *string 
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	IsDeleted bool
}

func (pg *PostgresDB) migrateUserTable()error{
	ddl := `
		CREATE TABLE IF NOT EXISTS users(
			id serial2 PRIMARY KEY,
			name varchar(20),
			email varchar(50) UNIQUE NOT NULL,
			email_verified boolean DEFAULT false,
			password varchar(50) NOT NULL,
			role varchar(10) NOT NULL,
			bio varchar(100),
			created_at timestamp,
			updated_at timestamp,
			is_del boolean DEFAULT false,
			deleted_at timestamp
		);
	`
	return pgx.BeginFunc(context.Background(),pg.Conn,func(tx pgx.Tx)error{
		_,err := tx.Exec(context.Background(),ddl)
		return err
	})

}

func (pg *PostgresDB) createUser(user *User)(uint,error){
	sql := `
		INSERT INTO users(name,email,email_verified,
		password,role,bio,created_at,updated_at,
		deleted_at,is_del) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id;
	`
	var id uint
	
	txopts := pgx.TxOptions{IsoLevel:"serializable"}

	tx,err := pg.Conn.BeginTx(context.Background(),txopts)
	if err != nil{
		return id,err
	}

	defer tx.Rollback(context.Background())

	err = tx.Query(context.Background(),sql,
		user.Name,user.Email,false,
		user.Password,user.Role,time.Now(),
		time.Now(),nil,false).Scan(&id)
	if err != nil{
		return id,err
	}
	
	if err := tx.Commit(context.Background()); err != nil{
		return id,err
	}

	return id,nil
}

func (pg *PostgresDB) checkExists(email string)(bool,error){
	sql := `
		SELECT COUNT(1) FROM users WHERE email = $1;
	`
	var count uint
	err := pg.Conn.Query(context.Background(),sql,email).Scan(&count)
	if err != nil{
		return false,err
	}

	return count == 0,nil
}


