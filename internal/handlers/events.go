package handlers

import (
	"log"
	"github.com/bwmarrin/discordgo"
)

func ReadyEventWrapper(logger *log.Logger) func(session *discordgo.Session, event *discordgo.Ready) {
	return func (session *discordgo.Session, event *discordgo.Ready) {
		logger.Printf("Connected as %v#%v", event.User.Username, event.User.Discriminator)
	}
}