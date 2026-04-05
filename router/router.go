package router

import (
	"Focogram/controllers"
	"Focogram/utils"

	"Focogram/middlewares"

	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func SetRouter() *gin.Engine {
	r := gin.Default()

	r.Use(CORSMiddleware())

	r.Static("/uploads", "./uploads")

	go utils.WsMgr.Start()
	public := r.Group("api")
	{
		public.POST("/register", controllers.Register)
		public.POST("/login", controllers.Login)
		public.POST("/password/reset", controllers.ResetPasswordByEmail)
		public.GET("/userinfo", controllers.GetUserInfo)
		public.GET("/post/:userid", controllers.GetUserPosts)
		public.GET("/post/liked/:userid", controllers.GetUserLikedPosts)
		public.GET("/post/detail/:postid", controllers.GetPostDetail)
		public.GET("/comment/:postid", controllers.GetPostComments)
		public.GET("/like/count/:postid", controllers.GetPostLikeCount2)
		public.GET("/user/following/:userid", controllers.GetUserFollowing)
		public.GET("/user/followers/:userid", controllers.GetUserFollowers)
	}

	auth := r.Group("api/auth")
	auth.Use(middlewares.AuthMiddleware())
	{
		auth.POST("/upload/avatar", controllers.UploadAvatar)
		auth.POST("/upload/image", controllers.UploadPostImage)
		auth.POST("/post", controllers.CreatePost)
		auth.DELETE("/post/:postid", controllers.DeletePost)
		auth.POST("/like/:postid", controllers.LikePost2)
		auth.DELETE("/like/:postid", controllers.LikePost2)
		auth.PATCH("/userinfo", controllers.UpdateUserInfo)
		auth.PATCH("/password", controllers.UpdatePassword)

		auth.GET("/like/:postid", controllers.GetPostLikeUsers2)

		auth.POST("/comment/:postid", controllers.CreateComment)
		auth.DELETE("/comment/:commentid", controllers.DeleteComment)

		auth.POST("/followuser/:userid", controllers.FollowUser)
		auth.POST("/unfollowuser/:userid", controllers.UnfollowUser)
		auth.GET("/checkfollow/:userid", controllers.CheckFollow)
		auth.GET("/following", controllers.GetMyFollowing)
		auth.GET("/followers", controllers.GetMyFollowers)
		auth.GET("/timeline", controllers.GetFollowingPosts)

		auth.GET("/notifications", controllers.GetNotifications)
		auth.POST("/notifications/read", controllers.MarkNotificationsAsRead)
		auth.DELETE("/notification/:id", controllers.DeleteNotification)
		auth.POST("/notifications/batch-delete", controllers.BatchDeleteNotifications)

		auth.POST("/message/conversation", controllers.CreateConversation)
		auth.POST("/message/send/:receiver_id/:conv_id", controllers.SendPrivateMessage)
		auth.GET("/message/conversation/:conv_id", controllers.GetConversationMessages)
		auth.GET("/message/conversations", controllers.GetConversations)
		auth.POST("/message/conversation/:conv_id/read", controllers.MarkConversationAsRead)
		auth.GET("/message/unread/stats", controllers.GetUnreadMessageStats)
	}

	r.GET("/ws", middlewares.AuthMiddleware(), utils.WsHandler)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "notification-service",
		})
	})

	return r
}
