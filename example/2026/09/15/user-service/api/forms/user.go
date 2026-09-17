package forms

type PasswordLoginForm struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=20"`
}

type RegisterForm struct {
	NickName string `form:"nick_name" json:"nick_name" binding:"required"`
	Mobile   string `form:"mobile" json:"mobile" binding:"required,len=11"`
	Password string `form:"password" json:"password" binding:"required,min=3,max=20"`
}
