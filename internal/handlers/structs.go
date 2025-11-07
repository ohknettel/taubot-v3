package handlers

import (
	"github.com/ohknettel/taubot-v3/internal/database"
	"github.com/bwmarrin/discordgo"
)

type Context struct {
	Backend database.Backend
	GetOptions func() []*discordgo.ApplicationCommandInteractionDataOption	
}

var Colors = struct{
	Error int
	Normal int
	Warning int
}{0xDC143C, 0x00ADD8, 0xE9D502}

type EventFunc func (self *Context, s *discordgo.Session, v *discordgo.InteractionCreate)

type Command struct {
	Name string
	Description string
	DefaultPermissions *int64
	Callback EventFunc
	Subcommands []*Command
	Options []*Option
}

type Option struct {
	Name string
	Description string
	Type any
	Required bool
	Autocomplete *EventFunc
	Choices []*discordgo.ApplicationCommandOptionChoice
	Options []*Option
}