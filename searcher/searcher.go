package searcher

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/pgvector/pgvector-go"
	"github.com/samber/lo"
	"github.com/xyenon/telemikiya/config"
	"github.com/xyenon/telemikiya/database"
	"github.com/xyenon/telemikiya/database/ent"
	entmessage "github.com/xyenon/telemikiya/database/ent/message"
	"github.com/xyenon/telemikiya/provider"
	"go.uber.org/fx"
)

type Params struct {
	fx.In

	Database          *database.Database
	EmbeddingProvider provider.EmbeddingProvider
	Config            *config.Config
}

type Searcher struct {
	db                *database.Database
	embeddingProvider provider.EmbeddingProvider
	cfg               *config.Config
}

func New(params Params) *Searcher {
	searcher := &Searcher{
		db:                params.Database,
		embeddingProvider: params.EmbeddingProvider,
		cfg:               params.Config,
	}

	return searcher
}

type SearchParams struct {
	fx.In

	Input     string    `name:"input"`
	Count     uint      `name:"count"`
	StartTime time.Time `name:"start_time"`
	EndTime   time.Time `name:"end_time"`
	DialogID  int64     `name:"dialog_id"`
}

func (s Searcher) Search(ctx context.Context, params SearchParams) ([]*ent.Message, error) {
	embeddings, err := s.embeddingProvider.Embed(ctx, []string{params.Input})
	if err != nil {
		return nil, fmt.Errorf("failed to embed messages: %w", err)
	}
	vector := pgvector.NewVector(embeddings[0])

	semanticSearch, fullTextSearch := "semantic_search", "full_text_search"
	fieldRank := "rank"
	dialectPostgres := sql.Dialect(dialect.Postgres)
	messageTable := dialectPostgres.Table(entmessage.Table)

	orderByEmbeddingFunc := func(b *sql.Builder) {
		b.WriteString(messageTable.C(entmessage.FieldTextEmbedding)).
			Pad().WriteString("<=>").Pad().
			WriteString("$1")
	}
	orderByPgroongaExpr := sql.Expr("pgroonga_score(tableoid, ctid)")
	coalesceBuilder := func(ident string) func(b *sql.Builder) {
		return func(b *sql.Builder) {
			b.WriteString("COALESCE").
				Wrap(func(b *sql.Builder) {
					b.WriteString(dialectPostgres.Table(semanticSearch).C(ident)).
						Comma().
						WriteString(dialectPostgres.Table(fullTextSearch).C(ident))
				})
		}
	}

	subQueryBuilder := func(mode string) sql.TableView {
		rankBuilder := sql.Window(func(b *sql.Builder) {
			b.WriteString("RANK").Wrap(func(b *sql.Builder) {})
		})

		q := dialectPostgres.Select(
			messageTable.C(entmessage.FieldID),
			messageTable.C(entmessage.FieldMsgID),
			messageTable.C(entmessage.FieldDialogID),
			messageTable.C(entmessage.FieldText),
			messageTable.C(entmessage.FieldSentAt),
		).From(messageTable).
			Limit(int(params.Count)).
			As(mode)

		switch mode {
		case semanticSearch:
			q = q.AppendSelectExprAs(
				rankBuilder.OrderExpr(sql.ExprP(
					dialectPostgres.String(orderByEmbeddingFunc),
					vector,
				)),
				fieldRank,
			)
			q = q.OrderExprFunc(orderByEmbeddingFunc)
		case fullTextSearch:
			q = q.AppendSelectExprAs(
				rankBuilder.OrderExpr(sql.DescExpr(orderByPgroongaExpr)),
				fieldRank,
			)
			q = q.OrderExpr(sql.DescExpr(orderByPgroongaExpr))
			q = q.Where(sql.P(func(b *sql.Builder) {
				b.WriteString(messageTable.C(entmessage.FieldText)).
					Pad().WriteString("&@*").Pad().
					Arg(params.Input)
			}))
		default:
			panic(fmt.Sprintf("unknown mode: %s", mode))
		}

		botIDStr := strings.Split(s.cfg.Telegram.BotToken, ":")[0]
		botID, err := strconv.ParseInt(botIDStr, 10, 64)
		if err == nil {
			q = q.Where(sql.NEQ(messageTable.C(entmessage.FieldDialogID), botID))
		}
		if !params.StartTime.IsZero() {
			q = q.Where(sql.GTE(messageTable.C(entmessage.FieldSentAt), params.StartTime))
		}
		if !params.EndTime.IsZero() {
			q = q.Where(sql.LTE(messageTable.C(entmessage.FieldSentAt), params.EndTime))
		}
		if lo.IsNotEmpty(params.DialogID) {
			q = q.Where(sql.EQ(messageTable.C(entmessage.FieldDialogID), params.DialogID))
		}

		return q
	}

	messages, err := s.db.Message.Query().
		Modify(func(s *sql.Selector) {
			s.Select().
				AppendSelectExprAs(dialectPostgres.Expr(coalesceBuilder(entmessage.FieldID)), entmessage.FieldID).
				AppendSelectExprAs(dialectPostgres.Expr(coalesceBuilder(entmessage.FieldMsgID)), entmessage.FieldMsgID).
				AppendSelectExprAs(dialectPostgres.Expr(coalesceBuilder(entmessage.FieldDialogID)), entmessage.FieldDialogID).
				AppendSelectExprAs(dialectPostgres.Expr(coalesceBuilder(entmessage.FieldText)), entmessage.FieldText).
				AppendSelectExprAs(dialectPostgres.Expr(coalesceBuilder(entmessage.FieldSentAt)), entmessage.FieldSentAt).
				AppendSelectExprAs(
					dialectPostgres.Expr(func(b *sql.Builder) {
						coalesceBuilder := func(table string) {
							b.WriteString("COALESCE").
								Wrap(func(b *sql.Builder) {
									b.WriteString(dialectPostgres.Table(table).C(fieldRank)).
										WriteOp(sql.OpDiv).
										WriteString(fmt.Sprintf("%.1f", float64(params.Count))).
										Comma().
										WriteString("0.0")
								})
						}
						coalesceBuilder(semanticSearch)
						b.WriteOp(sql.OpAdd)
						coalesceBuilder(fullTextSearch)
					}),
					fieldRank,
				).
				From(subQueryBuilder(semanticSearch)).
				FullJoin(subQueryBuilder(fullTextSearch)).
				On(
					dialectPostgres.Table(semanticSearch).C(entmessage.FieldID),
					dialectPostgres.Table(fullTextSearch).C(entmessage.FieldID),
				).
				OrderExprFunc(func(b *sql.Builder) {
					b.Ident(fieldRank).WriteString(" DESC")
				})
		}).
		Limit(int(params.Count)).
		WithDialog().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	return messages, nil
}
