package router

import (
	"Focogram/controllers"
	"Focogram/utils"

	"Focogram/middlewares"

	"net/http"

	"github.com/gin-gonic/gin"
)

func SetRouter() *gin.Engine {
	r := gin.Default()
	go utils.WsMgr.Start()
	public := r.Group("api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.GET("/userinfo", controllers.GetUserInfo)
		public.GET("/post/:userid", controllers.GetUserPosts)
		public.GET("/like/count/:postid", controllers.GetPostLikeCount2)
		public.GET("/comment/:postid", controllers.GetPostComments)

	}

	auth := r.Group("api/auth")
	auth.Use(middlewares.AuthMiddleware())
	{
		// 用户信息修改
		auth.PATCH("/userinfo", controllers.UpdateUserInfo)
		auth.PATCH("/password", controllers.UpdatePassword)

		// 帖子相关
		auth.POST("/post", controllers.CreatePost)
		auth.DELETE("/post/:postid", controllers.DeletePost)

		// 点赞相关
		auth.POST("/like/:postid", controllers.LikePost2)
		auth.GET("/like/:postid", controllers.GetPostLikeUsers2)

		// 评论相关
		auth.POST("/comment/:postid", controllers.CreateComment)
		auth.DELETE("/comment/:commentid", controllers.DeleteComment)

		// 关注相关
		auth.POST("/followuser/:userid", controllers.FollowUser)
		auth.POST("/unfollowuser/:userid", controllers.UnfollowUser)
		auth.GET("/following", controllers.GetMyFollowing)
		auth.GET("/followers", controllers.GetMyFollowers)

		// 通知相关
		auth.GET("/notifications", controllers.GetNotifications)

		//私信
		auth.POST("/message/conversation", controllers.CreateConversation)
		auth.POST("/message/send/:receiver_id/:conv_id", controllers.SendPrivateMessage)
		auth.GET("/message/conversation/:conv_id", controllers.GetConversationMessages)

		// WebSocket连接（实时消息推送）
		// 注意：WebSocket使用GET方法，且需要认证
		r.GET("/ws", middlewares.AuthMiddleware(), utils.WsHandler)

	}
	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "notification-service",
		})
	})

	return r

}
