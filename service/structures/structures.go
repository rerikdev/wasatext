package structures

type User struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	DisplayName    string `json:"displayName"`
	ProfilePicture string `json:"profilePicture"`
}

type Reaction struct {
	MessageId int    `json:"messageId"`
	UserId    int    `json:"userId"`
	Emoji     string `json:"emoji"`
	User      User   `json:"user"`
}

type Message struct {
	ID               int         `json:"id"`
	ConversationID   int         `json:"conversation_id"`
	Content          string      `json:"content"`
	IsForwarded      bool        `json:"isForwarded"`
	MediaType        string      `json:"mediaType"`
	Reactions        []*Reaction `json:"reactions"`
	Sender           User        `json:"sender"`
	Status           string      `json:"status"`
	Timestamp        string      `json:"timestamp"`
	ReplyToMessageID *int        `json:"replyToMessageId,omitempty"` // NEW: ID del messaggio a cui si risponde
	ReplyToMessage   *Message    `json:"replyToMessage,omitempty"`   // NEW: Oggetto messaggio completo a cui si risponde
}

type Conversation struct {
	ID      int    `json:"id"`
	Name    string `json:"name,omitempty"`
	Photo   string `json:"photo,omitempty"`
	IsGroup bool   `json:"is_group"`
	Members []User `json:"members"`
}

type Group struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Photo   string `json:"photo"`
	Members []User `json:"members"`
}

type SessionRequest struct {
	Name           string `json:"name"`
	DisplayName    string `json:"displayName,omitempty"`
	ProfilePicture string `json:"profilePicture,omitempty"`
}

type ConversationPreview struct {
	ID              int    `json:"id"`
	OtherUserID     int    `json:"otherUserId"`
	Username        string `json:"username"`
	ProfilePicture  string `json:"profilePicture"`
	LastMessage     string `json:"lastMessage"`
	LastMessageTime string `json:"lastMessageTime"`
}

type GroupPreview struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Photo           string `json:"photo"`
	Members         []User `json:"members"`
	LastMessage     string `json:"lastMessage"`
	LastMessageTime string `json:"lastMessageTime"`
}
