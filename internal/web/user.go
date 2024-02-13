package web

import (
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/pkg/ginx"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
	bizLogin             = "login"
)

type UserHandler struct {
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
	}
}
func (c *UserHandler) RegisterRoutes(server *gin.Engine) {
	// 分组注册
	ug := server.Group("/users")
	//ug.POST("/login", c.Login)
	ug.POST("/login", c.LoginJWT)
	ug.POST("/signup", c.SignUp)
	ug.POST("/edit", c.Edit)
	ug.POST("/login_sms/code/send", c.SendSMSLoginCode)
	ug.POST("/login_sms", c.LoginSMS)
	ug.GET("/profile", c.Profile)
}

// 登录
func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password)

	switch err {
	case nil:
		//写入session
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(
			sessions.Options{
				MaxAge: 900,
			})
		err = sess.Save()
		fmt.Println(err)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}

		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, service.ErrInvalidUserOrPassword.Error())
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}

// SignUp 注册
func (h *UserHandler) SignUp(ctx *gin.Context) {
	type SignupReq struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmpassword"`
	}
	var req SignupReq
	if err := ctx.Bind(&req); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		return
	}

	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不正确")
		return
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符、并且长度不能小于 8 位")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		ctx.String(http.StatusOK, "注册成功")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "注册失败")
	}
}

// Edit 修改
func (h *UserHandler) Edit(ctx *gin.Context) {
	type Req struct {
		// 邮箱 密码 手机号不允许在这个位置修改
		Nickname string `json:"nickname"`
		Birthday string `json:"birthday"`
		// 最大长度限制是128字节
		AboutMe string `json:"aboutme" binding:"max=128"`
	}

	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误,参数异常")
		return
	}
	//验证生日字段
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}

	uid, _ := ctx.Get("userId")
	u := domain.User{
		Id:       uid.(int64),
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	}

	err = h.svc.UpdateNonSensitiveInfo(ctx, u)

	if err != nil {
		ctx.String(http.StatusOK, "修改信息失败")
		return
	}

	ctx.String(http.StatusOK, "修改信息成功")
}

// 用户信息
func (h *UserHandler) Profile(ctx *gin.Context) {
	uc := ctx.MustGet("user").(UserClaims)
	u, err := h.svc.FindById(ctx, uc.Uid)

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	//返回给前端的数据
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		AboutMe  string `json:"aboutMe"`
		Birthday string `json:"birthday"`
	}

	resUser := User{
		Nickname: u.Nickname,
		Email:    u.Email,
		AboutMe:  u.AboutMe,
		Birthday: u.Birthday.Format(time.DateOnly),
	}

	ctx.JSON(200, resUser)

}

func (h *UserHandler) LoginJWT(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, req.Email, req.Password)

	switch err {
	case nil:
		h.setJWTToken(ctx, u.Id)
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, service.ErrInvalidUserOrPassword.Error())
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	err := ctx.Bind(&req)
	if err != nil {
		return
	}

	// 你这边可以校验 Req
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "请输入手机号码",
		})
		return
	}

	err = h.codeSvc.Send(ctx, bizLogin, req.Phone)
	if err != nil {
		return
	}
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		// 补日志的
	}
}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "验证码不对，请重新输入",
		})
		return
	}

	u, err := h.svc.FindOrCreate(ctx, req.Phone)

	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	h.setJWTToken(ctx, u.Id)
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "登录成功",
	})

}

func (h *UserHandler) setJWTToken(ctx *gin.Context, id int64) {
	uc := UserClaims{
		Uid:       id,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			// 1分钟过期token
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 30)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKEY)

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	//token 返回给前端
	ctx.Header("x-jwt-token", tokenStr)
	ctx.String(http.StatusOK, "登录成功")
}

var JWTKEY = []byte("00k2XQmvKq0uYdAtDy0msE6u6wpu8Fw0")

type UserClaims struct {
	Uid       int64
	UserAgent string
	jwt.RegisteredClaims
}
