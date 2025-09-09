package bot

import (
	"log"
	"github.com/bwmarrin/discordgo"
	"github.com/ohknettel/taubot-v3/internal/database"
	"github.com/ohknettel/taubot-v3/internal/handlers"
)

type Bot struct {
	Session *discordgo.Session
	Backend database.Backend
	Logger *log.Logger
}

func NewBot(token string) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	bot := Bot{Session: session}
	return &bot, nil
}

func (b *Bot) SetupBackend(uri string, driver database.DriverFunc, logger *log.Logger) error {
	db, err := handlers.PrepareDatabase(uri, driver, logger, 2) // 4 = info, 2 = error
	if err != nil {
		return err
	}

	backend := database.NewBackend(db)
	b.Backend = *backend
	return nil
}

func (b *Bot) SetLogger(logger *log.Logger) {
	b.Logger = logger
}


func (b *Bot) RegisterHandlers() error {
	dict, err := handlers.MigrateCommands(b.Session, Commands)
	if err != nil {
		return err
	}

	b.Session.AddHandler(handlers.SlashCommandHandlerWrapper(dict))
	b.Session.AddHandler(handlers.ReadyEventWrapper(b.Logger))
	return nil
}

func (b *Bot) Run() error {
	return b.Session.Open()
}