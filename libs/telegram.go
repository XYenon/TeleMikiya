package libs

import (
	"fmt"

	tdconstant "github.com/gotd/td/constant"
	"github.com/xyenon/telemikiya/database/ent"
	"github.com/xyenon/telemikiya/types"
)

// BotDialogToMTProto converts a bot API dialog to a MTProto dialog.
// https://core.telegram.org/api/bots/ids
func BotDialogToMTProto(botDialogID int64, dialogType types.DialogType) (int64, types.MTProtoDialogType) {
	switch dialogType {
	case types.TypeUser:
		return botDialogID, types.MTProtoDialogTypeUser
	case types.TypeGroup:
		if botDialogID > tdconstant.ZeroTDLibChannelID {
			return -botDialogID, types.MTProtoDialogTypeChat
		}
		fallthrough
	case types.TypeChannel:
		return tdconstant.ZeroTDLibChannelID - botDialogID, types.MTProtoDialogTypeChannel
	default:
		panic(fmt.Errorf("unknown dialog type: %s", dialogType))
	}
}

// DeepLink generates a Telegram deep link URL for the given message.
// https://core.telegram.org/api/links
func DeepLink(msg *ent.Message) string {
	dialog := msg.Edges.Dialog
	dialogID, dialogType := BotDialogToMTProto(dialog.ID, dialog.Type)
	switch dialogType {
	case types.MTProtoDialogTypeUser, types.MTProtoDialogTypeChat:
		return fmt.Sprintf("tg://user?id=%d", dialogID)
	case types.MTProtoDialogTypeChannel:
		return fmt.Sprintf("https://t.me/c/%d/%d", dialogID, msg.MsgID)
	default:
		panic(fmt.Errorf("unknown dialog type: %s", dialogType))
	}
}
