package internal

import(
"database/sql"
"github.com/jackc/pgx/v5"
"time"
"log"
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
		CREATE TABLE users IF NOT EXISTS(
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
	return pgx.BeginFunc(context.Background(),pg.conn,func(tx pgx.Tx)error{
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
	
	txopts := sql.TxOptions{Isolation:6}

	tx,err := pg.conn.BeginTx(context.Background(),&txopts)
	if err != nil{
		return err
	}

	defer tx.Rollback(context.Background())

	err := tx.Query(context.Background(),sql,
		user.Name,
		user.Email,
		false,
		user.Password,
		user.Role,
		time.Now(),
		time.Now(),
		nil,
		false).Scan(&id)
	if err != nil{
		return id,err
	}
	
	if err := tx.Commit(); err != nil{
		return id,err
	}

	return id,nil
}

//keeping email unique just for now
func (u *User) checkUserExists(email string)(bool,error){
	var exists bool
	
	t := time.Now()
	//levelserializable 
	txopts := sql.TxOptions{Isolation:6}
	dbc.pgdb.Transaction(func(db *gorm.DB){
		sql := `select count(1) from users where email = ?`
		r := db.Raw(sql,email).Scan(&exists)
		return r.Error
	},&txopts)

	log.Print("time taken: ",time.Since(t))
	
}

func (u *User) getById(id uint)(User){
	sql := `select name,email,role,created_at from users where id = ?`
	txopts := sql.TxOptions{Isolation:6}
	var name,email,role string
	var createdAt time.Time
	
	var errNotFound error

	err := dbc.pgdb.Transaction(func(db *gorm.DB){
		//error in executing sql
		row := db.Raw(sql,id).Row()
		if t := row.Err();t != nil{
			return t
		}
		r := row.Scan(&name,&email,&role,&createdAt)
		if r.Error == sql.ErrNoRows{
			errNotFound = errors.New("user doesnt exists.")
		}else{
			return r.Error
		}
	},&txopts)
	if err != nil{
		return nil,err
	}

	if errNotFound != nil{
		return nil,errNotFound
	}
	
	dbuser := User{Name:name,Email:email,Role:role,CreatedAt:createdAt}

	return &dbuser,nil
}


func (u *User) create()(uint,error){
	
	txopts := sql.TxOptions{Isolation:6}
	err := dbc.pgdb.Transaction(func(db *gorm.DB){
		res := db.Create(&u)
		return res.Error
	},&txopts)

	return u.ID,err
}

func (u *User) getPassword()(string,error){
	sql := `select password from users where name = ?`
	var dbpass string
	res := db.Raw(sql,u.name).Scan(&dbpass)
	return dbpass,res.Error
}

func (u *User) updateName()error{

}

func (u *User) updateEmail()error(){
	sql := "update users set email = ? where id = ?"
	txopts := &sql.TxOptions{Isolation:6}
	return dbc.pgdb.Transaction(func(db *gorm.DB){
		return db.Exec(sql,u.Email,u.ID)
	},&txopts)	
}

func (u *User) updatePassword()error{

}


func (u *User) emailVerified()error{
	sql := "update users set email_verified = true where id = ?"
	txopts := sql.TxOptions{Isolation:6}
	err := dbc.pgdb.Transaction(func(db *gorm.DB){
		return db.Exec(sql,u.ID)
	},&txopts)

	log.Print("time to verify email: ",time.Since(t))

	return err
}

func (u *User) delete()error{
	txopts := sql.TxOptions{Isolation:6}
	tx := dbc.pgdb.Begin(&txopts)
	defer func(){
		if r := recover();r != nil{
			log.Print("Fatal: ",r)
		}
	}
	if tx.Error != nil{
		return tx.Error
	}
	

	if r := tx.Delete(&User{},u.ID).Error; r != nil{
		tx.Rollback()
		return r
	}
	
	return tx.Commit().Error

}
