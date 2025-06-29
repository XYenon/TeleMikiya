// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	pgvector "github.com/pgvector/pgvector-go"
	"github.com/xyenon/telemikiya/database/ent/dialog"
	"github.com/xyenon/telemikiya/database/ent/message"
	"github.com/xyenon/telemikiya/types"
)

// MessageCreate is the builder for creating a Message entity.
type MessageCreate struct {
	config
	mutation *MessageMutation
	hooks    []Hook
}

// SetMsgID sets the "msg_id" field.
func (mc *MessageCreate) SetMsgID(i int) *MessageCreate {
	mc.mutation.SetMsgID(i)
	return mc
}

// SetDialogID sets the "dialog_id" field.
func (mc *MessageCreate) SetDialogID(i int64) *MessageCreate {
	mc.mutation.SetDialogID(i)
	return mc
}

// SetText sets the "text" field.
func (mc *MessageCreate) SetText(s string) *MessageCreate {
	mc.mutation.SetText(s)
	return mc
}

// SetTextEmbedding sets the "text_embedding" field.
func (mc *MessageCreate) SetTextEmbedding(pg pgvector.Vector) *MessageCreate {
	mc.mutation.SetTextEmbedding(pg)
	return mc
}

// SetNillableTextEmbedding sets the "text_embedding" field if the given value is not nil.
func (mc *MessageCreate) SetNillableTextEmbedding(pg *pgvector.Vector) *MessageCreate {
	if pg != nil {
		mc.SetTextEmbedding(*pg)
	}
	return mc
}

// SetHasMedia sets the "has_media" field.
func (mc *MessageCreate) SetHasMedia(b bool) *MessageCreate {
	mc.mutation.SetHasMedia(b)
	return mc
}

// SetMediaInfo sets the "media_info" field.
func (mc *MessageCreate) SetMediaInfo(ti *types.MediaInfo) *MessageCreate {
	mc.mutation.SetMediaInfo(ti)
	return mc
}

// SetSentAt sets the "sent_at" field.
func (mc *MessageCreate) SetSentAt(t time.Time) *MessageCreate {
	mc.mutation.SetSentAt(t)
	return mc
}

// SetID sets the "id" field.
func (mc *MessageCreate) SetID(u uuid.UUID) *MessageCreate {
	mc.mutation.SetID(u)
	return mc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (mc *MessageCreate) SetNillableID(u *uuid.UUID) *MessageCreate {
	if u != nil {
		mc.SetID(*u)
	}
	return mc
}

// SetDialog sets the "dialog" edge to the Dialog entity.
func (mc *MessageCreate) SetDialog(d *Dialog) *MessageCreate {
	return mc.SetDialogID(d.ID)
}

// Mutation returns the MessageMutation object of the builder.
func (mc *MessageCreate) Mutation() *MessageMutation {
	return mc.mutation
}

// Save creates the Message in the database.
func (mc *MessageCreate) Save(ctx context.Context) (*Message, error) {
	mc.defaults()
	return withHooks(ctx, mc.sqlSave, mc.mutation, mc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (mc *MessageCreate) SaveX(ctx context.Context) *Message {
	v, err := mc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mc *MessageCreate) Exec(ctx context.Context) error {
	_, err := mc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mc *MessageCreate) ExecX(ctx context.Context) {
	if err := mc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (mc *MessageCreate) defaults() {
	if _, ok := mc.mutation.ID(); !ok {
		v := message.DefaultID()
		mc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (mc *MessageCreate) check() error {
	if _, ok := mc.mutation.MsgID(); !ok {
		return &ValidationError{Name: "msg_id", err: errors.New(`ent: missing required field "Message.msg_id"`)}
	}
	if _, ok := mc.mutation.DialogID(); !ok {
		return &ValidationError{Name: "dialog_id", err: errors.New(`ent: missing required field "Message.dialog_id"`)}
	}
	if _, ok := mc.mutation.Text(); !ok {
		return &ValidationError{Name: "text", err: errors.New(`ent: missing required field "Message.text"`)}
	}
	if _, ok := mc.mutation.HasMedia(); !ok {
		return &ValidationError{Name: "has_media", err: errors.New(`ent: missing required field "Message.has_media"`)}
	}
	if _, ok := mc.mutation.MediaInfo(); !ok {
		return &ValidationError{Name: "media_info", err: errors.New(`ent: missing required field "Message.media_info"`)}
	}
	if _, ok := mc.mutation.SentAt(); !ok {
		return &ValidationError{Name: "sent_at", err: errors.New(`ent: missing required field "Message.sent_at"`)}
	}
	if len(mc.mutation.DialogIDs()) == 0 {
		return &ValidationError{Name: "dialog", err: errors.New(`ent: missing required edge "Message.dialog"`)}
	}
	return nil
}

func (mc *MessageCreate) sqlSave(ctx context.Context) (*Message, error) {
	if err := mc.check(); err != nil {
		return nil, err
	}
	_node, _spec := mc.createSpec()
	if err := sqlgraph.CreateNode(ctx, mc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	mc.mutation.id = &_node.ID
	mc.mutation.done = true
	return _node, nil
}

func (mc *MessageCreate) createSpec() (*Message, *sqlgraph.CreateSpec) {
	var (
		_node = &Message{config: mc.config}
		_spec = sqlgraph.NewCreateSpec(message.Table, sqlgraph.NewFieldSpec(message.FieldID, field.TypeUUID))
	)
	if id, ok := mc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := mc.mutation.MsgID(); ok {
		_spec.SetField(message.FieldMsgID, field.TypeInt, value)
		_node.MsgID = value
	}
	if value, ok := mc.mutation.Text(); ok {
		_spec.SetField(message.FieldText, field.TypeString, value)
		_node.Text = value
	}
	if value, ok := mc.mutation.TextEmbedding(); ok {
		_spec.SetField(message.FieldTextEmbedding, field.TypeOther, value)
		_node.TextEmbedding = value
	}
	if value, ok := mc.mutation.HasMedia(); ok {
		_spec.SetField(message.FieldHasMedia, field.TypeBool, value)
		_node.HasMedia = value
	}
	if value, ok := mc.mutation.MediaInfo(); ok {
		_spec.SetField(message.FieldMediaInfo, field.TypeJSON, value)
		_node.MediaInfo = value
	}
	if value, ok := mc.mutation.SentAt(); ok {
		_spec.SetField(message.FieldSentAt, field.TypeTime, value)
		_node.SentAt = value
	}
	if nodes := mc.mutation.DialogIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   message.DialogTable,
			Columns: []string{message.DialogColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dialog.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.DialogID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// MessageCreateBulk is the builder for creating many Message entities in bulk.
type MessageCreateBulk struct {
	config
	err      error
	builders []*MessageCreate
}

// Save creates the Message entities in the database.
func (mcb *MessageCreateBulk) Save(ctx context.Context) ([]*Message, error) {
	if mcb.err != nil {
		return nil, mcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(mcb.builders))
	nodes := make([]*Message, len(mcb.builders))
	mutators := make([]Mutator, len(mcb.builders))
	for i := range mcb.builders {
		func(i int, root context.Context) {
			builder := mcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*MessageMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, mcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, mcb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, mcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (mcb *MessageCreateBulk) SaveX(ctx context.Context) []*Message {
	v, err := mcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (mcb *MessageCreateBulk) Exec(ctx context.Context) error {
	_, err := mcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mcb *MessageCreateBulk) ExecX(ctx context.Context) {
	if err := mcb.Exec(ctx); err != nil {
		panic(err)
	}
}
