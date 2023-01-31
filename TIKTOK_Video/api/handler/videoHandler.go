// Code generated by hertz generator.

package handler

import (
	"TIKTOK_Video/model/vo"
	"TIKTOK_Video/mw"
	"TIKTOK_Video/service/ServiceImpl"
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	ResponseSuccess = 0
	ResponseFail    = 1
)

/*
视频流接口
*/

func Feed(ctx context.Context, c *app.RequestContext) {
	var userId int64 = -1
	var err error

	queryTime := c.Query("latest_time")
	token := c.Query("token")

	// 如果token不是空的，则可能登录状态，调用jwt鉴权。
	// 1. 鉴权失败，username为空，仍可以调用feed流
	// 2. 鉴权成功，可获取username
	// 如果token是空的，是未登录状态，username为空，仍可以调用feed流
	if token != "" {
		c.Next(ctx)
		if userid, exist := c.Get(mw.IdentityKey); exist {
			userId = userid.(int64)
		}
	}
	c.Abort()

	var latestTime int64
	if queryTime != "" {
		latestTime, err = strconv.ParseInt(queryTime, 10, 64)
	} else {
		latestTime = time.Now().UnixMilli()
	}

	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			StatusCode: ResponseFail,
			StatusMsg:  "时间戳请求错误",
		})
		return
	}
	vsi := ServiceImpl.VideoServiceImpl{}
	videoInfoList, nextTime, err := vsi.GetVideoInfosByLatestTime(latestTime, userId)
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			StatusCode: ResponseFail,
			StatusMsg:  "获取视频流失败：" + err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, vo.FeedResponse{
		Response:  vo.Response{StatusCode: ResponseSuccess},
		VideoList: videoInfoList,
		NextTime:  nextTime,
	})
}

// PublishAction
/*
	登录用户选择视频上传
*/
func PublishAction(ctx context.Context, c *app.RequestContext) {
	// get the basic info from meta
	userId, _ := c.Get(mw.IdentityKey)

	videoTitle := c.PostForm("title")
	videoData, err := c.Request.FormFile("data")
	if err != nil {
		log.Print("can not get this filestream")
	}
	vsi := ServiceImpl.VideoServiceImpl{}
	if err := vsi.PublishVideo(userId.(int64), videoData, videoTitle); err != nil {
		c.JSON(consts.StatusOK, vo.Response{
			StatusCode: ResponseFail,
			StatusMsg:  err.Error(),
		})
		return
	}

	c.JSON(consts.StatusOK, vo.Response{
		StatusCode: ResponseSuccess,
		StatusMsg:  "publish success!",
	})

}
