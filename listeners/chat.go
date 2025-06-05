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

	server.Client.OnPlayerConnect = append(server.Client.OnPlayerConnect, gbxclient.GbxCallbackStruct[events.PlayerConnectEventArgs]{
		Key:  "ChatListener",
		Call: cl.onPlayerConnect,
	})

	server.Client.OnPlayerDisconnect = append(server.Client.OnPlayerDisconnect, gbxclient.GbxCallbackStruct[events.PlayerDisconnectEventArgs]{
		Key:  "ChatListener",
		Call: cl.onPlayerDisconnect,
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

	// If the message starts with a slash, we don't need to handle the chat message
	if len(playerChatEvent.Text) > 0 && playerChatEvent.Text[0] == '/' {
		return
	}

	if cl.Server.Info.Chat.MessageFormat == "" {
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
	message := cl.Server.Info.Chat.MessageFormat.FormatMessage(
		player.Login,
		player.NickName,
		playerChatEvent.Text,
	)

	cl.Server.Client.ChatSendServerMessage(message)
}

func (cl *ChatListener) onPlayerConnect(playerConnectEvent events.PlayerConnectEventArgs) {
	if cl.Server.Info.Chat.ConnectMessage == "" {
		return
	}

	var player structs.PlayerInfo
	for _, p := range cl.Server.Info.ActivePlayers {
		if p.Login == playerConnectEvent.Login {
			player = p
			break
		}
	}

	message := cl.Server.Info.Chat.ConnectMessage.FormatMessage(
		player.Login,
		player.NickName,
		"",
	)

	cl.Server.Client.ChatSendServerMessage(message)
}

func (cl *ChatListener) onPlayerDisconnect(playerDisconnectEvent events.PlayerDisconnectEventArgs) {
	if cl.Server.Info.Chat.DisconnectMessage == "" {
		return
	}

	var player structs.PlayerInfo
	for _, p := range cl.Server.Info.ActivePlayers {
		if p.Login == playerDisconnectEvent.Login {
			player = p
			break
		}
	}

	message := cl.Server.Info.Chat.DisconnectMessage.FormatMessage(
		player.Login,
		player.NickName,
		"",
	)

	cl.Server.Client.ChatSendServerMessage(message)
}
