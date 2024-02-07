package dao

import (
	"context"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.CTime = now
	u.UTime = now
	err := dao.db.WithContext(ctx).Create(&u).Error

	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 唯一索引冲突
			return ErrDuplicateEmail
		}
	}

	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) UpdateById(ctx *gin.Context, entity User) error {
	// 这种写法依赖于 GORM 的零值和主键更新特性
	// Update 非零值 WHERE id = ?
	//return dao.db.WithContext(ctx).Updates(&entity).Error
	return dao.db.WithContext(ctx).Model(&entity).Where("id = ?", entity.Id).
		Updates(map[string]any{
			"u_time":   time.Now().UnixMilli(),
			"nickname": entity.Nickname,
			"birthday": entity.Birthday,
			"about_me": entity.AboutMe,
		}).Error
}

func (dao *UserDAO) FindById(ctx *gin.Context, uid int64) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id = ?", uid).First(&u).Error
	return u, err
}

func (dao *UserDAO) FindByPhone(ctx *gin.Context, phone string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("phone = ?", phone).First(&u).Error
	return u, err
}

type User struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 代表这是一个可以为 NULL 的列
	Email    sql.NullString `gorm:"unique"`
	Nickname string         `gorm:"type=varchar(128)"`
	Birthday int64
	AboutMe  string `gorm:"type=varchar(4096)"`
	Password string
	Phone    sql.NullString `gorm:"unique"`
	CTime    int64
	UTime    int64
}
