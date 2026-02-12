package api

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/rerikdev/wasatext/service/database"
)

type Router struct {
	Router *httprouter.Router
	db     database.AppDatabase
}

func New(db database.AppDatabase) (*Router, error) {
	r := httprouter.New()
	rt := &Router{Router: r, db: db}

	// Auth
	r.POST("/session", CORS(rt.doLogin))

	// Users
	r.GET("/users", CORS(rt.getUsers))
	r.PUT("/users/me/name", CORS(rt.setMyUserName))
	r.PUT("/users/me/photo", CORS(rt.setMyPhoto))

	// Conversations
	r.GET("/users/me/conversations", CORS(rt.getMyConversations))
	r.POST("/conversations", CORS(rt.createConversation))
	r.GET("/conversations/:conversationId", CORS(rt.getConversation))
	r.PUT("/conversations/:conversationId", CORS(rt.updateConversation))
	r.DELETE("/conversations/:conversationId", CORS(rt.deleteConversation))

	// Participants (group management)
	r.POST("/conversations/:conversationId/participants", CORS(rt.addParticipant))
	r.DELETE("/conversations/:conversationId/participants/:userId", CORS(rt.removeParticipant))
	r.GET("/conversations/:conversationId/participants", CORS(rt.getParticipants))

	// Messages
	r.POST("/conversations/:conversationId/messages", CORS(rt.sendMessage))
	r.DELETE("/conversations/:conversationId/messages/:messageId", CORS(rt.deleteMessage))

	// Reactions
	r.PUT("/conversations/:conversationId/messages/:messageId/reactions/:reactionId", CORS(rt.addReaction))
	r.DELETE("/conversations/:conversationId/messages/:messageId/reactions/:reactionId", CORS(rt.removeReaction))

	// Read receipts (for groups)
	r.POST("/conversations/:conversationId/read", CORS(rt.markMessagesAsRead))

	// Forwarding
	r.POST("/conversations/:conversationId/forward/:targetConversationId", CORS(rt.forwardMessage))

	return rt, nil
}

func (a *Router) Handler() http.Handler {
	return a.Router
}

// extractUser translates a URL parameter (which could be "Maria" OR "2") into an ID ("2")
func (rt *Router) extractUser(param string) (string, error) {
	id, err := rt.db.GetUserByName(param)
	if err != nil {
		return "", err
	}
	if id != "" {
		return id, nil
	}
	return param, nil
}