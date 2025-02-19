package record

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/celestix/gotgproto/ext"
	tgtypes "github.com/celestix/gotgproto/types"
	"github.com/gotd/td/tg"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"github.com/xyenon/telemikiya/database/ent"
	entdialog "github.com/xyenon/telemikiya/database/ent/dialog"
	"github.com/xyenon/telemikiya/types"
	"go.uber.org/zap"
)

func Recorder(ctx *ext.Context, update *ext.Update) error {
	logger := ctx.Value(types.CtxKeyLogger{}).(*zap.Logger)
	cfg := ctx.Value(types.CtxKeyConfigTelegram{}).(*config.Telegram)
	db := ctx.Value(types.ContextKeyDB{}).(*database.Database)

	dialog := types.FromEffectiveChat(update.EffectiveChat())
	dialogID, err := dialog.ID()
	if err != nil {
		return fmt.Errorf("failed to get dialog id: %w", err)
	}
	if len(cfg.ObservedDialogIDs) > 0 &&
		!lo.Contains(cfg.ObservedDialogIDs, dialogID) {
		logger.Info("dialog is not observed", zap.Int64("dialog_id", dialogID))
		return nil
	}

	err = saveDialog(ctx, logger, cfg, db, dialogID, update.EffectiveChat())
	if err != nil {
		err = fmt.Errorf("failed to save dialog: %w", err)
		return err
	}

	err = saveMessage(ctx, logger, db, dialogID, update.EffectiveMessage)
	if err != nil {
		err = fmt.Errorf("failed to save message: %w", err)
	}

	return nil
}

func saveDialog(
	ctx *ext.Context,
	logger *zap.Logger,
	cfg *config.Telegram,
	db *database.Database,
	dialogID int64,
	chat tgtypes.EffectiveChat,
) error {
	dialogLock := ctx.Value(types.CtxKeyDialogLock{}).(*sync.Map)
	dialog := types.FromEffectiveChat(chat)

	var locked bool
	for {
		if _, loaded := dialogLock.LoadOrStore(dialogID, struct{}{}); loaded {
			logger.Debug("dialog is locked, waiting", zap.Int64("dialog_id", dialogID))
			time.Sleep(time.Second)
		} else {
			logger.Debug("dialog is not locked", zap.Int64("dialog_id", dialogID))
			locked = true
			break
		}
	}
	defer func() {
		if locked {
			dialogLock.Delete(dialogID)
		}
	}()

	var exist, skip bool
	oldDialog, err := db.Dialog.Query().Where(entdialog.ID(dialogID)).Select(entdialog.FieldUpdatedAt).Only(ctx)
	switch {
	case ent.IsNotFound(err):
		// dialog does not exist, create new
		exist, skip = false, false
	case err != nil:
		// error occurred
		err = fmt.Errorf("failed to query dialog: %w", err)
		skip = true
	case oldDialog.UpdatedAt.Before(time.Now().Add(-cfg.DialogUpdateInterval)):
		// dialog exists but need to update
		exist, skip = true, false
	default:
		// dialog exists and no need to update
		exist, skip = true, true
		dialogLock.Delete(dialogID)
		locked = false
	}
	logger.Info("dialog status", zap.Int64("dialog_id", dialogID), zap.Bool("exist", exist), zap.Bool("skip", skip))
	if skip {
		return nil
	}

	dialogType, err := dialog.Type()
	if err != nil {
		return fmt.Errorf("failed to get dialog type: %w", err)
	}
	title, err := dialog.Title()
	if err != nil {
		return fmt.Errorf("failed to get dialog title: %w", err)
	}

	if exist {
		logger.Info("updating dialog", zap.Int64("dialog_id", dialogID), zap.String("title", title))
		_, err = db.Dialog.UpdateOneID(dialogID).SetTitle(title).Save(ctx)
	} else {
		logger.Info("creating dialog", zap.Int64("dialog_id", dialogID), zap.String("title", title))
		_, err = db.Dialog.Create().SetID(dialogID).SetTitle(title).SetType(dialogType).Save(ctx)
	}
	if err != nil {
		return fmt.Errorf("failed to save dialog: %w", err)
	}
	return nil
}

func saveMessage(
	ctx context.Context,
	logger *zap.Logger,
	db *database.Database,
	dialogID int64,
	msg *tgtypes.Message,
) (err error) {
	msgID, hasMedia, mediaInfo := msg.GetID(), false, types.MediaInfo{}
	sentAt := time.Unix(int64(msg.GetDate()), 0).UTC()
	media, ok := msg.GetMedia()
	if ok {
		switch v := media.(type) {
		case *tg.MessageMediaEmpty:
			logger.Debug("empty media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaPhoto:
			logger.Debug("photo media", zap.Int("msg_id", msgID))
			g, _ := v.GetPhoto()
			if p, ok := g.(*tg.Photo); ok {
				hasMedia = true
				mediaInfo.Type = "photo"
				mediaInfo.Photo = types.Photo{
					ID:                  p.GetID(),
					AccessHash:          p.GetAccessHash(),
					FileReferenceBase64: base64.StdEncoding.EncodeToString(p.GetFileReference()),
				}
			}
		case *tg.MessageMediaGeo:
			logger.Debug("geo media", zap.Int("msg_id", msgID))
			if g, ok := v.GetGeo().(*tg.GeoPoint); ok {
				hasMedia = true
				mediaInfo.Type = "geo"
				accuracyRadius, _ := g.GetAccuracyRadius()
				mediaInfo.GeoPoint = types.GeoPoint{
					Long:           g.GetLong(),
					Lat:            g.GetLat(),
					AccessHash:     g.GetAccessHash(),
					AccuracyRadius: accuracyRadius,
				}
			}
		case *tg.MessageMediaContact:
			logger.Debug("contact media", zap.Int("msg_id", msgID))
			hasMedia = true
			mediaInfo.Type = "contact"
			mediaInfo.Contact = types.Contact{
				PhoneNumber: v.GetPhoneNumber(),
				FirstName:   v.GetFirstName(),
				LastName:    v.GetLastName(),
				Vcard:       v.GetVcard(),
				UserID:      v.GetUserID(),
			}
		case *tg.MessageMediaUnsupported:
			logger.Debug("unsupported media", zap.Int("msg_id", msgID))
			hasMedia = true
			mediaInfo.Type = "unsupported"
		case *tg.MessageMediaDocument:
			logger.Debug("document media", zap.Int("msg_id", msgID))
			documentInfos := handleDocument(ctx, logger, v)
			if len(documentInfos) > 0 {
				hasMedia = true
				mediaInfo.Type = "documents"
			}
		case *tg.MessageMediaWebPage:
			logger.Debug("webpage media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaVenue:
			logger.Debug("venue media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaGame:
			logger.Debug("game media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaInvoice:
			logger.Debug("invoice media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaGeoLive:
			logger.Debug("geo live media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaPoll:
			logger.Debug("poll media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaDice:
			logger.Debug("dice media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaStory:
			logger.Debug("story media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaGiveaway:
			logger.Debug("giveaway media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaGiveawayResults:
			logger.Debug("giveaway results media", zap.Int("msg_id", msgID))
		case *tg.MessageMediaPaidMedia:
			logger.Debug("paid media", zap.Int("msg_id", msgID))
		default:
			logger.Warn("unknown media type", zap.Int("msg_id", msgID), zap.Any("media", media))
		}
	}

	logger.Info("saving message", zap.Int("msg_id", msgID), zap.Int64("dialog_id", dialogID))
	_, err = db.Message.Create().
		SetMsgID(msgID).
		SetDialogID(dialogID).
		SetText(msg.GetMessage()).
		SetHasMedia(hasMedia).
		SetMediaInfo(&mediaInfo).
		SetSentAt(sentAt).
		Save(ctx)
	if err != nil {
		err = fmt.Errorf("failed to save message: %w", err)
	}
	return
}

func handleDocument(
	ctx context.Context,
	logger *zap.Logger,
	m *tg.MessageMediaDocument,
) (documentInfos []types.Document) {
	gg, _ := m.GetAltDocuments()
	documentInfos = make([]types.Document, 0, len(gg))
	for _, g := range gg {
		if d, ok := g.(*tg.Document); ok {
			documentInfo := types.Document{
				ID:                  d.GetID(),
				AccessHash:          d.GetAccessHash(),
				FileReferenceBase64: base64.StdEncoding.EncodeToString(d.GetFileReference()),
				MimeType:            d.GetMimeType(),
				Type:                "file",
			}
			for _, attr := range d.GetAttributes() {
				switch v := attr.(type) {
				case *tg.DocumentAttributeAnimated:
					documentInfo.Type = "animated"
				case *tg.DocumentAttributeSticker:
					documentInfo.Type = "sticker"
				case *tg.DocumentAttributeVideo:
					documentInfo.Type = "video"
				case *tg.DocumentAttributeAudio:
					documentInfo.Type = "audio"
				case *tg.DocumentAttributeFilename:
					documentInfo.Filename = v.GetFileName()
				case *tg.DocumentAttributeCustomEmoji:
					documentInfo.Type = "custom_emoji"
				default:
					logger.Debug("unsupported document attribute", zap.Any("attr", attr))
				}
			}
			documentInfos = append(documentInfos, documentInfo)
		}
	}
	return
}
