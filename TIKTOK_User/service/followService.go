package service

import (
	"TIKTOK_User/model/vo"
	"TIKTOK_User/service/serviceImpl"
	"sync"
)

type FollowService interface {
	CreateNewRelation(userfromid, usertoid int64) (int64, error)
	DeleteRelation(userfromid, usertoid int64) error
	GetFollowListById(userId int64) ([]vo.UserInfo, error)
}

var (
	followservice FollowService

	followServiceOnce sync.Once
)

// NewCommentServiceInstance  单例模式返回service对象
func NewCommentServiceInstance() FollowService {
	followServiceOnce.Do(
		func() {
			followservice = &serviceImpl.FollowServiceImpl{}
		})
	return followservice
}

type followerService interface {
	GetFollowerListById(userId int64) ([]vo.UserInfo, error)
}

var (
	service2 followerService

	followerServiceOnce sync.Once
)

// NewCommentServiceInstance2 NewCommentServiceInstance  单例模式返回service对象
func NewCommentServiceInstance2() followerService {
	followServiceOnce.Do(
		func() {
			service2 = &serviceImpl.FollowerServiceImpl{}
		})
	return service2
}
