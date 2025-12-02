package handlers

import (
	"context"

	"whitelist-bot/internal/core"
	domainUser "whitelist-bot/internal/domain/user"
	domainWLRequest "whitelist-bot/internal/domain/wl_request"
	"whitelist-bot/internal/metastore"
	repository "whitelist-bot/internal/repository/wl_request"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type iUserRepository interface {
	UserByTelegramID(ctx context.Context, telegramID int64) (domainUser.User, error)
	UserByID(ctx context.Context, id domainUser.ID) (domainUser.User, error)
}

type iWLRequestRepository interface {
	CreateWLRequest(
		ctx context.Context,
		requesterID domainWLRequest.RequesterID,
		nickname domainWLRequest.Nickname,
	) (domainWLRequest.WLRequest, error)
	PendingWLRequests(ctx context.Context, limit int64) ([]domainWLRequest.WLRequest, error)
	PendingWLRequestsWithRequester(ctx context.Context, limit int64) ([]repository.PendingWLRequestWithRequester, error)
	WLRequestByID(ctx context.Context, id domainWLRequest.ID) (domainWLRequest.WLRequest, error)
	UpdateWLRequest(ctx context.Context, wlRequest domainWLRequest.WLRequest) (domainWLRequest.WLRequest, error)
}

type iMessageSender interface {
	SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error)
	AnswerCallbackQuery(ctx context.Context, params *bot.AnswerCallbackQueryParams) (bool, error)
	EditMessageText(ctx context.Context, params *bot.EditMessageTextParams) (*models.Message, error)
}

type botMessageSender struct {
	b *bot.Bot
}

// NewBotMessageSender creates iMessageSender from bot.Bot.
func NewBotMessageSender(b *bot.Bot) iMessageSender {
	return botMessageSender{b: b}
}

func (s botMessageSender) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	return s.b.SendMessage(ctx, params)
}

func (s botMessageSender) AnswerCallbackQuery(
	ctx context.Context,
	params *bot.AnswerCallbackQueryParams,
) (bool, error) {
	return s.b.AnswerCallbackQuery(ctx, params)
}

func (s botMessageSender) EditMessageText(
	ctx context.Context,
	params *bot.EditMessageTextParams,
) (*models.Message, error) {
	return s.b.EditMessageText(ctx, params)
}

type Handlers struct {
	userRepo      iUserRepository
	wlRequestRepo iWLRequestRepository
	metastore     metastore.Metastore
	sender        iMessageSender
	config        core.Config
}

func New(userRepo iUserRepository, wlRequestRepo iWLRequestRepository, metastore metastore.Metastore, config core.Config) *Handlers {
	return &Handlers{userRepo: userRepo, wlRequestRepo: wlRequestRepo, metastore: metastore, config: config}
}

func NewWithSender(
	userRepo iUserRepository,
	wlRequestRepo iWLRequestRepository,
	sender iMessageSender,
	config core.Config,
) *Handlers {
	return &Handlers{
		userRepo:      userRepo,
		wlRequestRepo: wlRequestRepo,
		sender:        sender,
		config:        config,
	}
}

func (h Handlers) botAnswerCallbackQuery(
	ctx context.Context,
	b *bot.Bot,
	params *bot.AnswerCallbackQueryParams,
) (bool, error) {
	if h.sender != nil {
		return h.sender.AnswerCallbackQuery(ctx, params)
	}
	return b.AnswerCallbackQuery(ctx, params)
}

func (h Handlers) botSendMessage(
	ctx context.Context,
	b *bot.Bot,
	params *bot.SendMessageParams,
) (*models.Message, error) {
	if h.sender != nil {
		return h.sender.SendMessage(ctx, params)
	}
	return b.SendMessage(ctx, params)
}

func (h Handlers) botEditMessageText(
	ctx context.Context,
	b *bot.Bot,
	params *bot.EditMessageTextParams,
) (*models.Message, error) {
	if h.sender != nil {
		return h.sender.EditMessageText(ctx, params)
	}
	return b.EditMessageText(ctx, params)
}
