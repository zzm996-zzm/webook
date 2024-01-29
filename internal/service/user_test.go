package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
) 

func TestPasswordEncrypt(t *testing.T){
	password := []byte("1234563#123456")
	//代价（cost） 越高，加密强度越高但是对cpu负载更大
	encrypted,err := bcrypt.GenerateFromPassword(password,bcrypt.DefaultCost)
	assert.NoError(t,err)
	// fmt.Println(string(encrypted))

	//测试解密
	err = bcrypt.CompareHashAndPassword(encrypted,[]byte("testting"))
	assert.NotNil(t,err)
	err = bcrypt.CompareHashAndPassword(encrypted,password)
	assert.NoError(t,err)
}