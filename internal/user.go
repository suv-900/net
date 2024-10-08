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
	FollowerCount uint
	FollowingCount uint
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
	IsDeleted bool
}

type UserRepo struct{
	conn *pgx.Conn	
}

func (u *UserRepo) Migrate(ctx context.Context)error{
	user_ddl := `
		CREATE TABLE IF NOT EXISTS users(
			id serial2 PRIMARY KEY,
			name varchar(20),
			email varchar(50) UNIQUE NOT NULL,
			email_verified boolean DEFAULT false,
			password varchar(50) NOT NULL,
			role varchar(10) NOT NULL,
			follower_count smallint,
			following_count smallint,
			created_at timestamp,
			updated_at timestamp,
			is_del boolean DEFAULT false,
			deleted_at timestamp
		);
		CREATE TABLE IF NOT EXISTS relationships(
			id serial2 PRIMARY KEY,
			follower_id	REFERENCES users(id) ON DELETE CASCADE, 
			following_id serial2 REFERENCES users(id) ON DELETE CASCADE 
		);
		CREATE TABLE IF NOT EXISTS user_info(
			id serial2 PRIMARY KEY,
			user_id serial2 REFERENCES users(id) ON DELETE CASCADE,
			bio varchar(100),
			pfp_url varchar(100)
		);
		
		CREATE OR REPLACE PROCEDURE followproc(follower_id int,following_id int)
		LANGUAGE plpgsql
		AS $$
		BEGIN
			INSERT INTO relationships(follower_id,following_id) 
			VALUES(follower_id,following_id);

			UPDATE users SET following_count=following_count+1 WHERE id = follower_id;
			UPDATE users SET follower_count=follower_count+1 WHERE id = following_id;
			
		END;$$;
		
		CREATE OR REPLACE PROCEDURE unfollowproc(follower_id int,following_id int)
		LANGUAGE plpgsql
		AS $$
		BEGIN 
			DELETE FROM relationships WHERE follower_id = follower_id AND following_id = following_id;	
			UPDATE users SET following_count=following_count-1 WHERE id = follower_id;
			UPDATE users SET follower_count=follower_count-1 WHERE id = following_id;

		END;$$;
			
	`
	tx,err := u.conn.BeginTx(ctx)
	defer tx.Rollback(ctx)

	if err != nil{
		log.Print("error while creating tx: ",err)
		return ErrInternalServerError
	}
	
	err = tx.Exec(ctx,ddl)
	if err != nil{
		log.Print("error migrating: ",err)
		return ErrInternalServerError 
	}

	err = tx.Commit(ctx)
	if err != nil{
		log.Print("error commiting: ",err)
		return ErrInternalServerError
	}
	
	return nil
}

func (u *UserRepo) CreateUser(user *User)(uint,error){
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
	
	txopts := pgx.TxOptions{IsoLevel:Serializable}

	tx,err := u.conn.BeginTx(context.Background(),txopts)
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

func (u *UserRepo) Follow(ctx context.Context,follower_id uint,following_id uint)error{
	sql := `CALL followproc(@follower_id,@following_id)`

	args := pgx.NamedArgs{
		"follower_id":follower_id,
		"following_id":following_id,
	}
	
	txopts := pgx.TxOptions{Isolation:pgx.Serializable}
	tx,err := u.conn.BeginTx(ctx)
	defer tx.Rollback(ctx)

	if err != nil{
		log.Print("error while creating tx: ",err)
		return ErrInternalServerError
	}
	
	err = tx.Exec(ctx,sql,args)
	if err != nil{
		log.Print("error executing sql: ",err)
		return ErrInternalServerError 
	}
	
	//double commit 
	err = tx.Commit(ctx)
	if err != nil{
		log.Print("error commiting: ",err)
		return ErrInternalServerError
	}
	
	return nil
}

func (u *UserRepo) Unfollow(ctx context.Context,follower_id uint,following_id uint)error{
	sql := `CALL unfollowproc(@follower_id,@following_id)`	
	args := pgx.NamedArgs{
		"follower_id":follower_id,
		"following_id":following_id,
	}
	
	txopts := pgx.TxOptions{Isolation:pgx.Serializable}
	tx,err := u.conn.BeginTx(ctx)
	defer tx.Rollback(ctx)

	if err != nil{
		log.Print("error while creating tx: ",err)
		return ErrInternalServerError
	}
	
	err = tx.Exec(ctx,sql,args)
	if err != nil{
		log.Print("error executing sql: ",err)
		return ErrInternalServerError 
	}

	err = tx.Commit(ctx)
	if err != nil{
		log.Print("error commiting: ",err)
		return ErrInternalServerError
	}
	
	return nil
}

func (u *UserRepo) GetFollowers(ctx context.Context,id,limit,offset uint)([]*User,error){
	sql := `
		

	`	
}

func (u *UserRepo) GetFollowing(ctx context.Context,id,limit,offset uint)([]*User,error){
	sql := `
		

	`	
}
func (u *UserRepo) Delete(ctx context.Context,id uint)error{
	sql := `UPDATE users SET is_del = true WHERE id = $1`
	
	txopts := pgx.TxOptions{IsoLevel:Serializable}
	tx,err := u.conn.BeginTx(ctx,txopts)
	defer tx.Rollback(ctx)

	if err != nil{
		log.Print("error while creating tx: ",err)
		return ErrInternalServerError
	}
	
	err = tx.Exec(ctx,sql,id)
	if err != nil{
		log.Print("error deleting user: ",err)
		return ErrInternalServerError 
	}

	err = tx.Commit(ctx)
	if err != nil{
		log.Print("error commiting: ",err)
		return ErrInternalServerError
	}

	return nil
}

func (u *UserRepo) DeleteForce(ctx context.Context,id uint)error{
	sql := `DELETE FROM users WHERE id = $1`
	
	txopts := pgx.TxOptions{IsoLevel:Serializable}
	tx,err := u.conn.BeginTx(ctx,txopts)
	defer tx.Rollback(ctx)

	if err != nil{
		log.Print("error while creating tx: ",err)
		return ErrInternalServerError
	}
	
	err = tx.Exec(ctx,sql,id)
	if err != nil{
		log.Print("error deleting(force) user: ",err)
		return ErrInternalServerError 
	}

	err = tx.Commit(ctx)
	if err != nil{
		log.Print("error commiting: ",err)
		return ErrInternalServerError
	}

	return nil
}

func (u *UserRepo) VerifyUserEmail(ctx context.Context,id uint)error{
	sql := `UPDATE users SET email_verified=true WHERE id = $1;`

	txopts := pgx.TxOptions{IsoLevel:Serializable}
	tx,err := u.conn.BeginTx(ctx,txopts)
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

func (u *UserRepo) CheckUserExists(ctx context.Context,email string)(bool,error){
	sql := `SELECT COUNT(1) FROM users WHERE email = $1;`
	
	txopts := pgx.TxOptions{IsoLevel:Serializable,AccessMode:ReadOnly}
	var count uint
	
	tx,err := u.conn.BeginTx(ctx,txopts)
	defer tx.RollBack(ctx)
	if err != nil{
		return false,err
	}
	
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

func (u *UserRepo) GetUser(ctx context.Context,id uint)(*User,error){
	sql := `
		SELECT name,email,email_verified,bio,role,created_at
		FROM users WHERE id = $1
	`
	user := User{}
	
	stmt,err := u.conn.Prepare(ctx,"getuser",sql)
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

func (u *UserRepo) GetPassword(ctx context.Context,id uint)(*string,error){
	sql := `SELECT password FROM users WHERE id = $1`

	txopts := pgx.TxOptions{IsoLevel:Serializable,AccessMode:ReadOnly}
	tx,err := u.conn.BeginTx(ctx,txopts)
	defer tx.Rollback(ctx)
	if err != nil{
		log.Print("error couldnt create transaction: ",err)
		return nil,err
	}
	var password string
	rows,err := tx.Query(ctx,sql,id)
	defer rows.Close()

	err = rows.Scan(&password)
	if err != nil{
		log.Print("error while scanning: ",err)
		return nil,err
	}
	
	if err := rows.Err();err != nil{
		log.Print("error in db: ",err)
		return nil,err
	}

	if err := tx.Commit();err != nil{
		log.Print("error while commit: ",err)
		return nil,err
	}
	
	return &password,nil
}


func (u *UserRepo) UpdatePassword(ctx context.Context,id uint,password string)error{
	sql := `UPDATE users SET password = @password WHERE id = @id;`
	txopts := pgx.TxOptions{IsoLevel:Serializable}
	tx,err := u.conn.BeginTx(ctx,txopts)
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

func (u *UserRepo) GetVerifiedUsers(ctx context.Context,limit,offset uint)([]*User,error){
	sql := `
		SELECT id,name,email,role FROM users WHERE email_verified=true
		ORDER BY created_at ASC LIMIT $1 OFFSET $2;
	`
	rows,err = u.conn.Query(ctx,sql,limit,offset)
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

func (u *UserRepo) GetDeletedUsers(ctx context.Context,limit,offset string)([]*User,error){
	sql := `
		SELECT id,name,email,role,created_at,deleted_at 
		FROM users WHERE is_del = true ORDER BY deleted_at ASC
		LIMIT $1 OFFSET $2;
	`
	rows,err := u.conn.Query(ctx,sql,limit,offset)
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

func (u *UserRepo) GetAllUsers(ctx context.Context,limit,offset uint,role string)([]*User,error){
	sql := `
		SELECT id,name,role,email,email_verified,created_at,updated_at
		WHERE role = @role ORDER BY created_at ASC LIMIT @limit OFFSET @offset
	`
	args := pgx.NamedArgs{
		"role":role,
		"limit":limit,
		"offset":offset,
	}
	
	rows,err := u.conn.Query(ctx,sql,args)
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
