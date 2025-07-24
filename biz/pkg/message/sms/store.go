package sms

import (
	"context"

	"github.com/byteflowing/base/biz/dal/model"
	"github.com/byteflowing/base/biz/dal/query"
	"github.com/byteflowing/base/kitex_gen/base"
)

type DbStore struct {
	db *query.Query
}

func NewDbStore(db *query.Query) *DbStore {
	return &DbStore{db: db}
}

func (d *DbStore) Save(ctx context.Context, message *model.MessageSms) error {
	return d.db.MessageSms.WithContext(ctx).Create(message)
}

func (d *DbStore) Update(ctx context.Context, message *model.MessageSms) error {
	q := d.db.MessageSms
	_, err := q.WithContext(ctx).Where(q.ID.Eq(message.ID)).Updates(message)
	return err
}

func (d *DbStore) GetSendingMessages(ctx context.Context, provider base.SmsProvider, limit uint32) ([]*model.MessageSms, error) {
	q := d.db.MessageSms
	return q.WithContext(ctx).Where(
		q.MsgStatus.Eq(int16(base.MessageStatus_MESSAGE_STATUS_SENDING)),
		q.Provider.Eq(int32(provider)),
	).Order(q.CreatedAt.Asc()).Limit(int(limit)).Find()
}

type EmptyStore struct{}

func NewEmptyStore() *EmptyStore {
	return &EmptyStore{}
}

func (e EmptyStore) Save(ctx context.Context, message *model.MessageSms) error {
	return nil
}

func (e EmptyStore) Update(ctx context.Context, message *model.MessageSms) error {
	return nil
}

func (e EmptyStore) GetSendingMessages(ctx context.Context, provider base.SmsProvider, limit uint32) ([]*model.MessageSms, error) {
	return nil, nil
}
