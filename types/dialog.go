package types

import (
	"fmt"

	tgtypes "github.com/celestix/gotgproto/types"
	tdconstant "github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/samber/lo"
)

type Dialog struct {
	dialog any
}

func FromEffectiveChat(chat tgtypes.EffectiveChat) Dialog {
	return Dialog{dialog: chat}
}

func FromPeerClass(peer tg.PeerClass) Dialog {
	return Dialog{dialog: peer}
}

// ID returns the bot API dialog ID.
// https://core.telegram.org/api/bots/ids
func (d Dialog) ID() (botDialogID int64, err error) {
	switch d := d.dialog.(type) {
	case *tgtypes.User:
		botDialogID = d.GetID()
	case *tg.PeerUser:
		botDialogID = d.GetUserID()

	case *tgtypes.Chat:
		botDialogID = -d.GetID()
	case *tg.PeerChat:
		botDialogID = -d.GetChatID()

	case *tgtypes.Channel:
		botDialogID = tdconstant.ZeroTDLibChannelID - d.GetID()
	case *tg.PeerChannel:
		botDialogID = tdconstant.ZeroTDLibChannelID - d.GetChannelID()

	default:
		err = fmt.Errorf("unknown dialog type: %T %#v", d, d)
	}
	return
}

func (d Dialog) Type() (dialogType DialogType, err error) {
	switch d := d.dialog.(type) {
	case *tgtypes.User:
		dialogType = TypeUser
	case *tg.PeerUser:
		dialogType = TypeUser

	case *tgtypes.Chat:
		dialogType = TypeGroup
	case *tg.PeerChat:
		dialogType = TypeGroup

	case *tgtypes.Channel:
		if d.Raw().GetMegagroup() || d.Raw().GetGigagroup() {
			dialogType = TypeGroup
		} else {
			dialogType = TypeChannel
		}
	case *tg.PeerChannel:
		err = fmt.Errorf("peer channel is not supported: %T %#v", d, d)

	default:
		err = fmt.Errorf("unknown dialog type: %T %#v", d, d)
	}
	return
}

func (d Dialog) Title() (title string, err error) {
	switch d := d.dialog.(type) {
	case *tgtypes.User:
		title = fmt.Sprintf("%s %s", d.FirstName, d.LastName)
		if lo.IsNotEmpty(d.Username) {
			title += fmt.Sprintf(" (@%s)", d.Username)
		}
	case *tgtypes.Chat:
		title = d.Title
	case *tgtypes.Channel:
		title = d.Title
		if lo.IsNotEmpty(d.Username) {
			title += fmt.Sprintf(" (@%s)", d.Username)
		}

	case *tg.PeerClass:
		err = fmt.Errorf("peer class is not supported: %T %#v", d, d)

	default:
		err = fmt.Errorf("unknown dialog type: %T %#v", d, d)
	}
	return
}

type DialogType string

const (
	TypeUser    DialogType = "user"
	TypeGroup   DialogType = "group"
	TypeChannel DialogType = "channel"
)

// Values provides list valid values for Enum.
func (DialogType) Values() (kinds []string) {
	for _, s := range []DialogType{TypeUser, TypeGroup, TypeChannel} {
		kinds = append(kinds, string(s))
	}
	return
}

type MTProtoDialogType string

const (
	MTProtoDialogTypeUser    MTProtoDialogType = "user"
	MTProtoDialogTypeChat    MTProtoDialogType = "chat"
	MTProtoDialogTypeChannel MTProtoDialogType = "channel"
)
