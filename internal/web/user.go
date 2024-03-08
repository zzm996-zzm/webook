package web

import (
	"errors"
	"fmt"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	ijwt "webook/internal/web/jwt"
	"webook/pkg/ginx"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,72}$`
	bizLogin             = "login"
)

type UserHandler struct {
	ijwt.Handler
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, hdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		Handler:        hdl,
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
	}
}
func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// 分组注册
	ug := server.Group("/users")
	//ug.POST("/login", c.Login)
	ug.POST("/login", h.LoginJWT)

	//TODO: 所有的都修改
	ug.POST("/signup", ginx.WrapBody(h.SignUp))

	ug.POST("/edit", h.Edit)
	ug.POST("/login_sms/code/send", h.SendSMSLoginCode)
	ug.POST("/login_sms", h.LoginSMS)
	ug.GET("/profile", h.Profile)
	ug.POST("/logout", h.LogoutJWT)
	ug.GET("/refresh_token", h.RefreshToken)
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
func (h *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {

	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return ginx.Result{
			Code: 4,
			Msg:  "非法邮箱格式",
		}, nil
	}

	if req.Password != req.ConfirmPassword {
		ctx.String(http.StatusOK, "两次密码不正确")
		return ginx.Result{
			Code: 4,
			Msg:  "两次输入的密码不相等",
		}, nil
	}

	isPassword, err := h.passwordRexExp.MatchString(req.Password)

	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	if !isPassword {
		ctx.String(http.StatusOK, "密码必须包含数字、特殊字符、并且长度不能小于 8 位")
		return ginx.Result{
			Code: 4,
			Msg:  "密码必须包含字母、数字、特殊字符",
		}, nil
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch {
	case err == nil:
		return ginx.Result{
			Msg: "OK",
		}, nil
	case errors.Is(err, service.ErrDuplicateEmail):
		return ginx.Result{
			Code: 200,
			Msg:  "邮箱冲突",
		}, nil
	default:
		return ginx.Result{
			Msg: "系统错误",
		}, err
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

	uc := ctx.MustGet("user").(ijwt.UserClaims)
	u := domain.User{
		Id:       uc.Uid,
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

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	// 约定，前端在 Authorization 里面带上这个 refresh_token
	tokenStr := h.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RCJWTKey, nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// token 无效或者 redis 有问题
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{Code: 5, Msg: "系统错误"})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "退出登录成功"})
}

// 用户信息
func (h *UserHandler) Profile(ctx *gin.Context) {
	uc := ctx.MustGet("user").(ijwt.UserClaims)
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

		err = h.SetLoginToken(ctx, u.Id)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
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
	err = h.SetLoginToken(ctx, u.Id)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "登录成功",
	})

}
