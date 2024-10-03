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
	Bio string 
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
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

func (pg *PostgresDB) checkUserExists(ctx context.Context,email string)(bool,error){
	sql := `SELECT COUNT(1) FROM users WHERE email = $1;`
	txopts := pgx.TxOptions{IsoLevel:"serializable",AccessMode:"read only"}
	var count uint
	
	tx,err := pg.Conn.BeginTx(ctx,txopts)
	if err != nil{
		return false,err
	}
	
	defer tx.RollBack(ctx)

	rows,err := tx.Query(ctx,sql,email)
	
	defer rows.Close()
	
	if err := rows.Scan(&count) err != nil{
		log.Print("scan error: ",err)
		return false,err
	}

	if rows.Err() != nil{
		log.Print("rows error: ",rows.Err())
		return false,err
	}
	
	if err := tx.Commit();err != nil{
		log.Print("commit error: ",err)
		return false,err
	}

	return count == 0,nil
}

func (pg *PostgresDB) getUser(ctx context.Context,id uint)(*User,error){
	sql := `
		SELECT name,email,email_verified,bio,role,created_at
		FROM users WHERE id = $1
	`
	user := User{}
	
	stmt,err := pg.Conn.Prepare(ctx,"getuser",sql)
	defer stmt.Close()
	if err != nil{
		log.Print("error preparing statement: ",err)
		return nil,err
	}
	
	err = stmt.QueryRow(id).Scan(&user.Name,&user.Email,&user.EmailVerified,
		&user.Bio,&user.Role,&user.CreatedAt)
	
	if err != nil{
		if err == sql.ErrNoRows{
			return nil,errors.New("user not found.")
		}
		return nil,err	
	}

	return &user,nil
}

func (pg *PostgresDB) getVerifiedUsersList()([]User,error){
	sql := `
		SELECT name,email,
	`
}












