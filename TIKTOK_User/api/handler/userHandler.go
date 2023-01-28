// Code generated by hertz generator.

package handler

import (
	"TIKTOK_User/model/vo"
	"TIKTOK_User/mw"
	"TIKTOK_User/service/serviceImpl"
	"context"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
)

const (
	ResponseSuccess = 0
	ResponseFail    = 1
)

// UserInfo
/*
	用户信息接口，获取登录用户的id、昵称，如果实现社交部分的功能，还会返回关注数和粉丝数
*/
func UserInfo(ctx context.Context, c *app.RequestContext) {
	// 查询对象的userId
	queryUserId := c.Query("user_id")
	if queryUserId == "" {
		c.JSON(consts.StatusOK, vo.UserInfoResponse{
			Response: vo.Response{StatusCode: ResponseFail, StatusMsg: "query user_id empty"},
		})
		return
	}
	id, _ := strconv.ParseInt(queryUserId, 10, 64)

	// 通过token获取到的登录用户名
	userId, _ := c.Get(mw.IdentityKey)

	// 查询昵称、关注数、粉丝数
	usi := serviceImpl.UserServiceImpl{}
	if u, err := usi.GetUserInfoById(id, userId.(int64)); err == nil {
		c.JSON(consts.StatusOK, vo.UserInfoResponse{
			Response: vo.Response{StatusCode: ResponseSuccess},
			UserInfo: u,
		})
	} else {
		c.JSON(consts.StatusOK, vo.UserInfoResponse{
			Response: vo.Response{StatusCode: ResponseFail, StatusMsg: "error :" + err.Error()},
		})
	}
}

// Register
/*
	用户注册接口，新用户注册时提供用户名，密码，昵称即可，用户名需要唯一。创建成功后，返回用户id和权限token
*/
func Register(ctx context.Context, c *app.RequestContext) {
	username := c.Query("username")
	password := c.Query("password")
	if username == "" || password == "" {
		c.JSON(consts.StatusOK, vo.RegisterResponse{
			Response: vo.Response{StatusCode: ResponseFail, StatusMsg: "query username or password empty"},
		})
		c.Abort()
		return
	}
	usi := serviceImpl.UserServiceImpl{}
	if _, err := usi.CreateUserByNameAndPassword(username, password); err != nil {
		c.JSON(consts.StatusOK, vo.RegisterResponse{
			Response: vo.Response{
				StatusCode: ResponseFail,
				StatusMsg:  "error :" + err.Error()},
		})
		c.Abort()
	}
}
