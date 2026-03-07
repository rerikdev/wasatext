package api

import (
	"net/http"
)

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))
	rt.router.POST("/session", rt.doLogin)

	// Profile routes
	rt.router.GET("/users/:userId", rt.getUser)
	rt.router.PATCH("/users/:userId", rt.setMyUserName)
	rt.router.PATCH("/users/:userId/photo", rt.setMyPhoto)
	rt.router.GET("/search/users", rt.searchUsers)

	rt.router.POST("/conversations", rt.createConversation)
	rt.router.POST("/conversations/:id/messages", rt.sendMessage)
	rt.router.GET("/conversations/:id/messages", rt.getConversation)
	rt.router.GET("/conversations", rt.getMyConversations)
	rt.router.PATCH("/conversations/:id/messages/read", rt.markMessagesRead)
	rt.router.DELETE("/conversations/:id/messages/:messageId", rt.deleteMessage)
	rt.router.POST("/conversations/:id/messages/:messageId/forward", rt.forwardMessage)
	//rt.router.POST("/conversations/:id/messages/:messageId/reactions", rt.commentMessage)
	//rt.router.DELETE("/conversations/:id/messages/:messageId/reactions", rt.uncommentMessage)
	//rt.router.GET("/conversations/:id/messages/:messageId/reactions", rt.getMessageReactions)
	rt.router.POST("/groups", rt.addToGroup)
	rt.router.GET("/groups", rt.listGroups)
	rt.router.DELETE("/groups/:id/members", rt.leaveGroup)
	rt.router.PATCH("/groups/:id/name", rt.setGroupName)
	rt.router.PATCH("/groups/:id/photo", rt.setGroupPhoto)
	rt.router.PATCH("/groups/:id/members", rt.addGroupMembers)

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
