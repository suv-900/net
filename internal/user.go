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
		deleted_at,is_del) VALUES(@name,@email,@email_verified,@password
		@role,@bio,@createdat,@updatedat,@deleted_at,@is_del)
		RETURNING id;
	`
	args := pgx.NamedArgs{
		"name":user.Name,
		"email":user.Email,
		"email_verified":false,
		"password":user.Password,
		"role":user.Role,
		"bio":user.Bio,
		"createdat":time.Now(),
		"updatedat":time.Now(),
		"deletedat":nil,
		"is_del":false,
	}

	var id uint
	
	txopts := pgx.TxOptions{IsoLevel:"serializable"}

	tx,err := pg.Conn.BeginTx(context.Background(),txopts)
	defer tx.Rollback(context.Background())
	
	if err != nil{
		return id,err
	}

	err = tx.Query(context.Background(),sql,args).Scan(&id)
	if err != nil{
		return id,err
	}
	
	if err := tx.Commit(context.Background()); err != nil{
		return id,err
	}

	return id,nil
}

func (pg *PostgresDB) verifyUser(ctx context.Context,id uint)error{
	sql := `UPDATE users SET email_verified=true WHERE id = $1;`

	txopts := pgx.TxOptions{IsoLevel:"serializable"}
	tx,err := pg.Conn.BeginTx(ctx,txopts)
	defer tx.Rollback(ctx)
	
	if err != nil{
		log.Print("error creating transaction: ",err)
	}

	stmt,err := tx.Prepare(ctx,"verify_email",sql)
	defer stmt.Close()

	if err != nil{
		log.Print("error preparing sql: ",err)
		return err
	}

	err = stmt.Exec(ctx,id)
	if err != nil{
		log.Print("error while executing sql: ",err)
		return err
	}

	if err := tx.Commit(); err != nil{
		log.Print("error while making commits: ",err)
		return err
	}

	return nil
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

func (pg *PostgresDB) updatePassword(ctx context.Context,id uint,password string)error{
	sql := `
		UPDATE users SET password = @password
		WHERE id = @id;
	`
	txopts := pgx.TxOptions{IsoLevel:"serializable"}
	tx,err := pg.Conn.BeginTx(ctx,txopts)
	defer tx.Rollback(ctx)
	if err != nil{
		log.Print("error while creating transaction: ",err)
		return err
	}
	
	_,err := tx.Exec(ctx,sql,pgx.NamedArgs{"password":password,"id":id})
	if err != nil{
		log.Print("error occured during executing update: ",err)
		return err
	}
	
	if err := tx.Commit();err != nil{
		log.Print("error occured while commit: ",err)
		return err
	}
	
	return nil
}

func (pg *PostgresDB) getVerifiedUsersList(ctx context.Context,limit,offset uint)([]*User,error){
	sql := `
		SELECT id,name,email,role FROM users WHERE email_verified=true
		ORDER BY created_at ASC LIMIT $1 OFFSET $2;
	`
	rows,err = pg.Conn.Query(ctx,sql,limit,offset)
	defer rows.Close()
	
	var users []*User
	while rows.Next(){
		var user User
		err = rows.Scan(&user.Id,&user.Name,&user.Email,&user.Role)
		if err != nil{
			log.Print("error occured in db: ",err)
			return nil,err
		}
		users = append(users,&user)
	}
	
	if err := rows.Err();err != nil{
		log.Print("error occured in db: ",err)
		return nil,err
	}
	
	return users,nil
}

func (pg *PostgresDB) getDeletedUsers(ctx context.Context,limit,offset string)([]*User,error){
	sql := `
		SELECT id,name,email,role,created_at,deleted_at 
		FROM users WHERE is_del = true ORDER BY deleted_at ASC
		LIMIT $1 OFFSET $2;
	`
	rows,err := pg.Conn.Query(ctx,sql,limit,offset)
	defer rows.Close()
	
	var users []*User
	while rows.Next(){
		var user User
		
		err = rows.Scan(&user.Id,&user.Name,&user.Email,&user.Role,&user.CreatedAt,&user.DeletedAt)
		if err != nil{
			log.Print("error occured while scanning: ",err)
			return nil,err
		}

		users = append(users,&user)
	}
	
	if err := rows.Err();err != nil{
		log.Print("error occured in db: ",err)
		return nil,err
	}

	return users,nil
}

func (pg *PostgresDB) getAllUsers(ctx context.Context,limit,offset uint,role string)([]*User,error){
	sql := `
		SELECT id,name,role,email,email_verified,created_at,updated_at
		WHERE role = @role ORDER BY created_at ASC LIMIT @limit OFFSET @offset
	`
	args := pgx.NamedArgs{
		"role":role,
		"limit":limit,
		"offset":offset,
	}
	
	rows,err := pg.Conn.Query(ctx,sql,args)
	defer rows.Close()
	
	var users []*User
	while rows.Next(){
		var user User
		err = rows.Scan(&user.Id,&user.Name,&user.Role,&user.Email,&user.EmailVerified,
		&user.CreatedAt,&user.UpdatedAt)
		if err != nil{
			log.Print("error while scanning rows: ",err)
			return nil,err
		}
		users = append(users,&user)
	}
	if err := rows.Err();err != nil{
		log.Print("error occured in db: ",err)
		return nil,err
	}

	return users,nil
}


