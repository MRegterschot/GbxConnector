package listeners

import (
	"github.com/MRegterschot/GbxConnector/structs"
	"github.com/MRegterschot/GbxRemoteGo/events"
	"github.com/MRegterschot/GbxRemoteGo/gbxclient"
)

type ChatListener struct {
	Server *structs.Server
}

func AddChatListeners(server *structs.Server) *ChatListener {
	cl := &ChatListener{Server: server}
	server.Client.OnPlayerChat = append(server.Client.OnPlayerChat, gbxclient.GbxCallbackStruct[events.PlayerChatEventArgs]{
		Key:  "ChatListener",
		Call: cl.onPlayerChat,
	})
	return cl
}

func (cl *ChatListener) onPlayerChat(playerChatEvent events.PlayerChatEventArgs) {
	// If manual routing is not enabled, we don't need to handle the chat message
	if !cl.Server.Info.Chat.ManualRouting {
		return
	}

	// If the login is empty, we don't need to handle the chat message
	if playerChatEvent.Login == "" {
		return
	}

	if cl.Server.Info.Chat.OverrideFormat == "" {
		// If no override format is set, just send the raw message to everyone
		cl.Server.Client.ChatForwardToLogin(playerChatEvent.Text, playerChatEvent.Login, "")
		return
	}

	var player structs.PlayerInfo
	for _, p := range cl.Server.Info.ActivePlayers {
		if p.Login == playerChatEvent.Login {
			player = p
			break
		}
	}

	// Format the message using the override format
	message := cl.Server.Info.Chat.OverrideFormat.FormatMessage(
		player.Login,
		player.NickName,
		playerChatEvent.Text,
	)

	cl.Server.Client.ChatSendServerMessage(message)
}
