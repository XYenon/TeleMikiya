package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	atlasmigrate "ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
	atlasschema "ariga.io/atlas/sql/schema"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database/ent"
	entmessage "github.com/xyenon/telemikiya/database/ent/message"
	"github.com/xyenon/telemikiya/database/ent/migrate"
	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

type Params struct {
	fx.In

	LifeCycle           fx.Lifecycle
	Config              *config.Config
	Logger              *zap.Logger
	Debug               bool `name:"debug" optional:"true"`
	AllowClearEmbedding bool `name:"allowClearEmbedding" optional:"true"`
}

type Database struct {
	embeddingDimensions uint
	allowClearEmbedding bool
	logger              *zap.Logger
	UserSessionConn     *sql.DB
	BotSessionConn      *sql.DB
	*ent.Client
}

func New(params Params) (*Database, error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%d sslmode=%s connect_timeout=%d user=%s password='%s' dbname=%s",
		params.Config.Database.Host,
		params.Config.Database.Port,
		params.Config.Database.SSLMode,
		uint(params.Config.Database.ConnectTimeout.Seconds()),
		params.Config.Database.User,
		params.Config.Database.Password,
		params.Config.Database.DBName,
	)

	userSessionConn, err := sql.Open("postgres", dataSourceName+" search_path='user_session'")
	if err != nil {
		return nil, fmt.Errorf("failed to open user session database connection: %w", err)
	}
	botSessionConn, err := sql.Open("postgres", dataSourceName+" search_path='bot_session'")
	if err != nil {
		return nil, fmt.Errorf("failed to open bot session database connection: %w", err)
	}

	opts := make([]ent.Option, 0)
	if params.Debug {
		opts = append(opts, ent.Debug(), ent.Log(params.Logger.Sugar().Debug))
	}
	entClient, err := ent.Open("postgres", dataSourceName+" search_path='\"$user\",public,vectors'", opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to open ent client: %w", err)
	}

	db := &Database{
		embeddingDimensions: params.Config.Embedding.Dimensions,
		allowClearEmbedding: params.AllowClearEmbedding,
		logger:              params.Logger,
		UserSessionConn:     userSessionConn,
		BotSessionConn:      botSessionConn,
		Client:              entClient,
	}

	if params.LifeCycle != nil {
		params.LifeCycle.Append(fx.Hook{
			OnStop: func(ctx context.Context) error {
				return db.Close()
			},
		})
	}

	return db, nil
}

func (d *Database) Close() error {
	return errors.Join(
		d.UserSessionConn.Close(),
		d.BotSessionConn.Close(),
		d.Client.Close(),
	)
}

func (d *Database) Migrate(ctx context.Context) error {
	return d.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
		schema.WithDiffHook(func(next schema.Differ) schema.Differ {
			return schema.DiffFunc(func(current, desired *atlasschema.Schema) ([]atlasschema.Change, error) {
				changes, err := next.Diff(current, desired)
				if err != nil {
					return nil, err
				}

				embeddingDimensionsChanges := d.diffEmbeddingDimensions(current, desired)
				changes = append(changes, embeddingDimensionsChanges...)
				d.modifyEmbeddingDimensions(changes)

				return changes, nil
			})
		}),
		schema.WithApplyHook(func(next schema.Applier) schema.Applier {
			return schema.ApplyFunc(func(ctx context.Context, conn dialect.ExecQuerier, plan *atlasmigrate.Plan) error {
				if d.isEmbeddingDimensionsChanged(plan) {
					if !d.allowClearEmbedding {
						return ErrNotAllowedToClearEmbedding
					}
					d.logger.Info("embedding dimensions changed, need to clear text embedding")
					err := d.Message.Update().ClearTextEmbedding().Exec(ctx)
					if err != nil {
						return fmt.Errorf("failed to clear text embedding: %w", err)
					}
				}

				return next.Apply(ctx, conn, plan)
			})
		}),
	)
}

func (d *Database) diffEmbeddingDimensions(current, desired *atlasschema.Schema) (changes []atlasschema.Change) {
	currentTable, ok := lo.Find(current.Tables,
		func(t *atlasschema.Table) bool { return t.Name == entmessage.Table },
	)
	if !ok {
		return
	}
	currentCol, ok := lo.Find(currentTable.Columns,
		func(c *atlasschema.Column) bool { return c.Name == entmessage.FieldTextEmbedding },
	)
	if !ok {
		return
	}
	desiredTable, ok := lo.Find(desired.Tables,
		func(t *atlasschema.Table) bool { return t.Name == entmessage.Table },
	)
	if !ok {
		return
	}
	desiredCol, ok := lo.Find(desiredTable.Columns,
		func(c *atlasschema.Column) bool { return c.Name == entmessage.FieldTextEmbedding },
	)
	if !ok {
		return
	}

	currentColType := currentCol.Type.Type.(*postgres.UserDefinedType)
	desiredColType := desiredCol.Type.Type.(*postgres.UserDefinedType)
	realDesiredColType := fmt.Sprintf(desiredColType.T, d.embeddingDimensions)
	if currentColType.T != realDesiredColType {
		d.logger.Info("embedding dimensions changed", zap.String("from", currentColType.T), zap.String("to", realDesiredColType))

		tableChanges := []atlasschema.Change{
			&atlasschema.ModifyColumn{
				From:   currentCol,
				To:     desiredCol,
				Change: atlasschema.ChangeType,
			},
		}
		currentIndex, currentIndexOk := lo.Find(currentTable.Indexes,
			func(i *atlasschema.Index) bool { return i.Name == "message_text_embedding" },
		)
		desiredIndex, desiredIndexOk := lo.Find(desiredTable.Indexes,
			func(i *atlasschema.Index) bool { return i.Name == "message_text_embedding" },
		)
		if currentIndexOk && desiredIndexOk {
			tableChanges = append(tableChanges,
				&atlasschema.DropIndex{I: currentIndex},
				&atlasschema.AddIndex{I: desiredIndex},
			)
		}

		changes = []atlasschema.Change{
			&atlasschema.ModifyTable{
				T:       currentTable,
				Changes: tableChanges,
			},
		}
	}
	return
}

func (d *Database) modifyEmbeddingDimensions(changes []atlasschema.Change) {
	for _, c := range changes {
		switch c := c.(type) {
		case *atlasschema.AddTable:
			if c.T.Name != entmessage.Table {
				continue
			}
			for _, col := range c.T.Columns {
				if col.Name != entmessage.FieldTextEmbedding {
					continue
				}
				t := col.Type.Type.(*postgres.UserDefinedType)
				t.T = fmt.Sprintf(t.T, d.embeddingDimensions)
			}
		case *atlasschema.ModifyTable:
			if c.T.Name != entmessage.Table {
				continue
			}
			for _, cc := range c.Changes {
				var col *atlasschema.Column

				switch cc := cc.(type) {
				case *atlasschema.AddColumn:
					col = cc.C
				case *atlasschema.ModifyColumn:
					col = cc.To
				default:
					continue
				}
				if col.Name != entmessage.FieldTextEmbedding {
					continue
				}

				t := col.Type.Type.(*postgres.UserDefinedType)
				t.T = fmt.Sprintf(t.T, d.embeddingDimensions)
			}
		}
	}
}

func (d *Database) isEmbeddingDimensionsChanged(plan *atlasmigrate.Plan) bool {
	for _, c := range plan.Changes {
		cc, ok := c.Source.(*atlasschema.ModifyTable)
		if !ok || cc.T.Name != entmessage.Table {
			continue
		}
		for _, ccc := range cc.Changes {
			cccc, ok := ccc.(*atlasschema.ModifyColumn)
			if ok &&
				cccc.To.Name == entmessage.FieldTextEmbedding &&
				cccc.Change == atlasschema.ChangeType &&
				cccc.From.Type.Type.(*postgres.UserDefinedType).T != cccc.To.Type.Type.(*postgres.UserDefinedType).T {
				return true
			}
		}
	}
	return false
}
