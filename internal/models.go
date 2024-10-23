package internal

import(
"context"
"os"
"log"
)
type Models struct{
	User interface{
		Migrate(c context.Context)error
		
		CreateUser(c context.Context,u *User)(uint,error)
		VerifyUserEmail(c context.Context,email string)error
		CheckUserExists(c context.Context,email string)(bool,error)
		GetUser(c context.Context,id uint)(*User,error)
		GetPassword(c context.Context,email string)(*string,error)
		UpdatePassword(c context.Context,id uint,password string)error
		Delete(c context.Context,id uint)error
		DeleteForce(c context.Context,id uint)error
		Follow(c context.Context,follower_id uint,following_id)error
		Unfollow(c context.Context,follower_id uint,following_id)error
		
		GetFollowers(c context.Context,id,limit,offset uint)([]*User,error)
		GetFollowing(c context.Context,id,limit,offset uint)([]*User,error)
		GetVerifiedUsers(c context.Context,limit,offset uint)([]*User,error)
		GetDeletedUsers(c context.Context,limit,offset uint)([]*User,error)
		GetAllUsers(c context.Context,limit,offset uint)([]*User,error)
	}
}

const(
	ErrInternalServerError = errors.New("server error.")
	ErrConflict = errors.New("already exists.")
)

func CreateModel()(*Models,error){
	
	dburl,exists := os.LookupEnv("dburl")
	if !exists{
		log.Print("env var dburl doesnt exists")
		return nil,ErrInternalServerError
	}
	
	conn,err := CreatePostgresConn(context.Background(),dburl)
	if err != nil{
		return nil,ErrInternalServerError
	}
	
	models := &Models{
		User: UserRepo{conn:conn}
	}
	
	if  err := models.User.Migrate(context.Background()); err != nil{
		return nil,ErrInternalServerError
	}

	return models,nil
}

