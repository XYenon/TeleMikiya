package embedding

import (
	"context"
	"time"

	"github.com/samber/lo"
	"github.com/xyenon/pgvectors-go"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"github.com/xyenon/telemikiya/database/ent"
	entmessage "github.com/xyenon/telemikiya/database/ent/message"
	"github.com/xyenon/telemikiya/embedding/provider"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	LifeCycle         fx.Lifecycle
	Config            *config.Config
	Logger            *zap.Logger
	Database          *database.Database
	EmbeddingProvider provider.Provider
}

type Embedding struct {
	cfg               *config.Embedding
	logger            *zap.Logger
	db                *database.Database
	embeddingProvider provider.Provider

	ctx    context.Context
	cancel context.CancelFunc
}

func New(params Params) *Embedding {
	ctx, cancel := context.WithCancel(context.Background())

	embedding := &Embedding{
		cfg:               &params.Config.Embedding,
		logger:            params.Logger,
		db:                params.Database,
		embeddingProvider: params.EmbeddingProvider,
		ctx:               ctx,
		cancel:            cancel,
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go embedding.Run()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				embedding.cancel()
				return nil
			},
		})
	}

	return embedding
}

func (e *Embedding) Run() {
	for {
		select {
		case <-e.ctx.Done():
			return
		default:
		}

		messages, err := e.db.Message.Query().
			Select(entmessage.FieldID, entmessage.FieldText).
			Where(entmessage.TextEmbeddingIsNil()).
			Limit(int(e.cfg.BatchSize)).
			All(e.ctx)
		if err != nil {
			e.logger.Error("failed to query messages", zap.Error(err))
			continue
		}
		e.logger.Info("fetched messages", zap.Int("count", len(messages)))
		if len(messages) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		messageTexts := lo.Map(messages, func(msg *ent.Message, _ int) string { return msg.Text })
		embeddings, err := e.embeddingProvider.Embed(e.ctx, messageTexts)
		if err != nil {
			e.logger.Error("failed to embed messages", zap.Error(err))
			continue
		}

		for i, message := range messages {
			e.logger.Info("saving embedding", zap.String("text", message.Text))
			_, err = message.Update().SetTextEmbedding(pgvectors.NewVector(embeddings[i])).Save(e.ctx)
			if err != nil {
				e.logger.Error("failed to save embedding", zap.Error(err))
			}
		}
	}
}

func (e *Embedding) Stop() {
	e.cancel()
	if err := e.embeddingProvider.Close(); err != nil {
		e.logger.Error("failed to close embedding provider", zap.Error(err))
	}
	e.logger.Info("embedding service has been stopped")
}
