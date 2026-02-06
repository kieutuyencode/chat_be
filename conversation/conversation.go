package conversation

import (
	"backend/apperror"
	"backend/common/result"
	"backend/database/ent"
	"backend/database/ent/conversation"
	"backend/database/ent/conversationmember"
	"backend/database/ent/message"
	"backend/database/ent/user"
	"backend/database/predicate"
	"backend/file"
	"backend/http/pagination"
	"backend/websocket"
	"context"
	"strconv"

	"entgo.io/ent/dialect/sql"
	"github.com/cockroachdb/errors"
	"go.uber.org/fx"
)

type Conversation struct {
	file      file.File
	websocket *websocket.Websocket
}

type conversationParams struct {
	fx.In
	File      file.File
	Websocket *websocket.Websocket
}

func newConversation(p conversationParams) *Conversation {
	return &Conversation{
		file:      p.File,
		websocket: p.Websocket,
	}
}

func (s *Conversation) GetOnlineUsers(ctx context.Context, client *ent.Client, p *GetOnlineUsersParams) ([]*ent.User, error) {
	users, err := client.User.Query().
		Where(user.IsActive(true), user.IDNEQ(p.UserId)).
		Order(ent.Asc(user.FieldLastActiveAt)).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "User.Query() failed")
	}

	return users, nil
}

func (s *Conversation) Load(ctx context.Context, client *ent.Client, p *LoadParams) (*ent.Conversation, error) {
	if p.FromUserId == p.ToUserId {
		return nil, apperror.BadRequest("Unable to create a conversation with yourself", nil, nil)
	}

	c, err := client.Conversation.Query().
		Where(func(s *sql.Selector) {
			t := sql.Table(conversationmember.Table)
			s.Join(t).On(s.C(conversation.FieldID), t.C(conversationmember.FieldConversationId)).
				Where(sql.InInts(t.C(conversationmember.FieldUserId), []int{p.FromUserId, p.ToUserId}...)).
				GroupBy(s.C(conversation.FieldID)).
				Having(sql.P(func(b *sql.Builder) {
					b.WriteString("count(").
						Ident(conversationmember.FieldUserId).
						WriteString(") >= 2")
				}))
		}).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, errors.Wrap(err, "Conversation.Query() failed")
	}
	if c != nil {
		return c, nil
	}

	c, err = client.Conversation.Create().
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Conversation.Create() failed")
	}

	cMembers, err := client.ConversationMember.CreateBulk(
		client.ConversationMember.Create().
			SetUserID(p.FromUserId).
			SetConversationID(c.ID),
		client.ConversationMember.Create().
			SetUserID(p.ToUserId).
			SetConversationID(c.ID),
	).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "ConversationMember.CreateBulk() failed")
	}

	c.Edges.Members = cMembers

	return c, nil
}

func (s *Conversation) Get(ctx context.Context, client *ent.Client, p *GetParams) (*GetResult, error) {
	paginateQuery := &pagination.Query{
		Limit: p.Limit,
		Page:  p.Page,
	}
	limit, offset := paginateQuery.LimitOffset()

	res := &GetResult{
		Limit: paginateQuery.Limit,
		Page:  paginateQuery.Page,
	}

	queryBuilder := client.Conversation.
		Query().
		Where(
			conversation.HasMembersWith(
				conversationmember.UserIdEQ(p.UserId),
			),
		).
		WithMembers(func(q *ent.ConversationMemberQuery) {
			q.WithUser(func(q *ent.UserQuery) {
				q.Select(user.FieldFullname, user.FieldEmail, user.FieldAvatar, user.FieldIsActive, user.FieldLastActiveAt)
			})
			q.Where(
				conversationmember.HasUserWith(
					user.IDNEQ(p.UserId),
				),
			)
		}).
		WithMessages(func(q *ent.MessageQuery) {
			q.Modify(func(s *sql.Selector) {
				s.SelectExpr(
					sql.P(func(b *sql.Builder) {
						b.WriteString("DISTINCT ON(").Ident(s.C(message.FieldConversationId)).
							WriteString(") ").
							Ident(s.C(message.FieldID)).
							Comma().
							Ident(s.C(message.FieldCreatedAt)).
							Comma().
							Ident(s.C(message.FieldContent)).
							Comma().
							Ident(s.C(message.FieldIsSeen)).
							Comma().
							Ident(s.C(message.FieldConversationId)).
							Comma().
							Ident(s.C(message.FieldUserId))

					}),
				)
				s.OrderBy(s.C(message.FieldConversationId), sql.Desc(s.C(message.FieldCreatedAt)))
			}).WithMedia()
		})
	if p.Search != "" {
		queryBuilder.Where(
			conversation.HasMembersWith(
				conversationmember.HasUserWith(
					user.And(
						user.IDNEQ(p.UserId), // bỏ chính mình ra
						user.Or(
							predicate.UnaccentContainsFold(user.FieldFullname, p.Search),
							user.EmailContainsFold(p.Search),
						),
					),
				),
			),
		)
	}

	count, err := queryBuilder.Clone().Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Count failed")
	}
	res.Count = count

	rows, err := queryBuilder.
		Limit(limit).
		Offset(offset).
		Order(func(s *sql.Selector) {
			t := sql.Table(message.Table)
			s.LeftJoin(t).On(s.C(conversation.FieldID), t.C(message.FieldConversationId))
			s.GroupBy(s.C(conversation.FieldID))
			s.OrderExpr(sql.P(func(b *sql.Builder) {
				b.WriteString("MAX(").
					Ident(t.C(message.FieldCreatedAt)).
					WriteString(") DESC NULLS LAST")
			}))
			s.OrderBy(sql.Desc(s.C(conversation.FieldID)))
		}).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "All failed")
	}

	// ---- Query unread counts ----
	ids := make([]int, len(rows))
	for i, c := range rows {
		ids[i] = c.ID
	}

	unreadMap := make(map[int]int)
	type unreadRow struct {
		ConversationId int `json:"conversation_id"`
		Count          int `json:"count"`
	}
	var unreadRows []unreadRow

	err = client.Message.Query().
		Where(
			message.ConversationIdIn(ids...),
			message.UserIdNEQ(p.UserId),
			message.IsSeen(false),
		).
		GroupBy(message.FieldConversationId).
		Aggregate(ent.Count()).
		Scan(ctx, &unreadRows)
	if err != nil {
		return nil, errors.Wrap(err, "Unread count query failed")
	}
	for _, ur := range unreadRows {
		unreadMap[ur.ConversationId] = ur.Count
	}

	// ---- Map to DTO ----
	res.Rows = make([]*ConversationResponse, len(rows))
	for i, c := range rows {
		cr := &ConversationResponse{
			Conversation: c,
			UnreadCount:  unreadMap[c.ID],
		}

		res.Rows[i] = cr
	}

	// Total unread count
	totalUnreadCount, err := client.Message.Query().
		Where(
			message.UserIdNEQ(p.UserId),
			message.IsSeen(false),
			message.HasConversationWith(
				conversation.HasMembersWith(
					conversationmember.UserIdEQ(p.UserId),
				),
			),
		).
		Count(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Total unread count query failed")
	}
	res.TotalUnreadCount = totalUnreadCount

	return res, nil
}

func (s *Conversation) GetOne(ctx context.Context, client *ent.Client, p *GetOneParams) (*ent.Conversation, error) {
	res, err := client.Conversation.
		Query().
		Where(
			conversation.ID(p.ConversationId),
			conversation.HasMembersWith(
				conversationmember.UserIdEQ(p.UserId),
			),
		).
		WithMembers(func(q *ent.ConversationMemberQuery) {
			q.WithUser(func(q *ent.UserQuery) {
				q.Select(user.FieldFullname, user.FieldEmail, user.FieldAvatar, user.FieldIsActive, user.FieldLastActiveAt)
			})
			q.Where(
				conversationmember.HasUserWith(
					user.IDNEQ(p.UserId),
				),
			)
		}).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, errors.Wrap(err, "GetOne failed")
	}
	if res == nil {
		return nil, apperror.NotFound("Data not found", nil, nil)
	}

	return res, nil
}

func (s *Conversation) ValidateUserInConversation(ctx context.Context, client *ent.Client, p *ValidateUserInConversationParams) error {
	exists, err := client.Conversation.
		Query().
		Where(
			conversation.ID(p.ConversationId),
			conversation.HasMembersWith(
				conversationmember.UserIdEQ(p.UserId),
			),
		).
		Exist(ctx)
	if err != nil {
		return errors.Wrap(err, "Query failed")
	}
	if !exists {
		return apperror.BadRequest("You are not in the conversation", nil, nil)
	}

	return nil
}

func (s *Conversation) GetMessage(ctx context.Context, client *ent.Client, p *GetMessageParams) (*pagination.Result[*ent.Message], error) {
	err := s.ValidateUserInConversation(ctx, client, &ValidateUserInConversationParams{
		UserId:         p.UserId,
		ConversationId: p.ConversationId,
	})
	if err != nil {
		return nil, err
	}

	//Seen message
	err = client.Message.Update().
		Where(
			message.ConversationId(p.ConversationId),
			message.UserIdNEQ(p.UserId),
			message.IsSeen(false),
		).
		SetIsSeen(true).
		Exec(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Update failed")
	}

	// Send event seen message
	members, err := client.ConversationMember.
		Query().
		Where(
			conversationmember.ConversationId(p.ConversationId),
			conversationmember.UserIdNEQ(p.UserId),
		).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Query failed")
	}
	for _, member := range members {
		s.websocket.Clients().Group(
			strconv.Itoa(member.UserId)).
			Send(websocket.EventMessageSeen,
				result.Success("", ""),
			)
	}

	queryBuilder := client.Message.
		Query().
		Where(
			message.ConversationId(p.ConversationId),
		).
		WithMedia().
		Order(ent.Desc(message.FieldCreatedAt))

	res, err := pagination.Paginate(ctx, queryBuilder, &pagination.Query{
		Limit: p.Limit,
		Page:  p.Page,
	})
	if err != nil {
		return nil, errors.Wrap(err, "GetMessage failed")
	}

	return res, nil
}

func (s *Conversation) CreateMessage(ctx context.Context, client *ent.Client, p *CreateMessageParams) (*ent.Message, error) {
	err := s.ValidateUserInConversation(ctx, client, &ValidateUserInConversationParams{
		UserId:         p.UserId,
		ConversationId: p.ConversationId,
	})
	if err != nil {
		return nil, err
	}

	message, err := client.Message.Create().
		SetUserID(p.UserId).
		SetConversationID(p.ConversationId).
		SetContent(p.Content).
		Save(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Message.Create failed")
	}

	var builders []*ent.MessageMediaCreate
	for _, mediaData := range p.Media {
		mediaSrc, err := s.file.MoveFromTemporary(mediaData.Src, file.FolderMessageMedia)
		if err != nil {
			return nil, err
		}

		builders = append(builders, client.MessageMedia.Create().
			SetSrc(mediaSrc).
			SetMessageID(message.ID))
	}

	res, err := client.MessageMedia.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "CreateBulk failed")
	}

	message.Edges.Media = res

	// Send message to all members in conversation
	members, err := client.ConversationMember.
		Query().
		Where(
			conversationmember.ConversationId(p.ConversationId),
			conversationmember.UserIdNEQ(p.UserId),
		).
		All(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Query failed")
	}
	for _, member := range members {
		s.websocket.Clients().Group(
			strconv.Itoa(member.UserId)).
			Send(websocket.EventMessageReceived,
				result.Success("", message),
			)
	}

	return message, nil
}

type LoadParams struct {
	FromUserId int
	ToUserId   int
}

type GetParams struct {
	UserId int
	Limit  int
	Page   int
	Search string
}

type GetResult struct {
	Count            int                     `json:"count"`
	Rows             []*ConversationResponse `json:"rows"`
	Limit            int                     `json:"limit"`
	Page             int                     `json:"page"`
	TotalUnreadCount int                     `json:"totalUnreadCount"`
}

type ConversationResponse struct {
	Conversation *ent.Conversation `json:"conversation"`
	UnreadCount  int               `json:"unreadCount"`
}

type GetOneParams struct {
	UserId         int
	ConversationId int
}

type ValidateUserInConversationParams struct {
	UserId         int
	ConversationId int
}

type GetMessageParams struct {
	UserId         int
	ConversationId int
	Limit          int
	Page           int
}

type CreateMessageParams struct {
	UserId         int
	ConversationId int
	Content        string
	Media          []*CreateMedia
}

type CreateMedia struct {
	Src string
}

type GetOnlineUsersParams struct {
	UserId int
}
