package handlers

import (
	"os"
	"slices"

	"github.com/bwmarrin/discordgo"
)

func MigrateCommands(session *discordgo.Session, commands []Command) (map[string]Command, error) {
	var dict map[string]Command = make(map[string]Command, len(commands))
	for _, cmd := range commands {
		command := &discordgo.ApplicationCommand{Name: cmd.Name, Description: cmd.Description, DefaultMemberPermissions: cmd.DefaultPermissions, Options: ConvertOptions(cmd.Options)}
		compact(cmd, &CommandWrapper{command})

		dict[cmd.Name] = cmd
		_, err := session.ApplicationCommandCreate(session.State.User.ID, os.Getenv("GUILD_ID"), command)
		if err != nil {
			return nil, err
		}
	}
	return dict, nil
}

// Wrapper that takes a map of string->Command values and returns a handler for a discordgo.Session's InteractionCreate event; whereas the map is generated dynamically and needed for command navigation
func SlashCommandHandlerWrapper(commands map[string]Command) func(session *discordgo.Session, event *discordgo.InteractionCreate) {
	return func (session *discordgo.Session, event *discordgo.InteractionCreate) {
		ctx := Context{}

		switch event.Type {
		case discordgo.InteractionApplicationCommand:
			options := event.ApplicationCommandData().Options
			if h, ok := commands[event.ApplicationCommandData().Name]; ok {
				cmd, opt := TraverseCommand(h, options)
				ctx.GetOptions = func() []*discordgo.ApplicationCommandInteractionDataOption {
					return opt
				}

				cmd.Callback(&ctx, session, event)
			}

		case discordgo.InteractionApplicationCommandAutocomplete:
			options := event.ApplicationCommandData().Options
			if h, ok := commands[event.ApplicationCommandData().Name]; ok {
				cmd, opt := TraverseCommand(h, options)

				ctx.GetOptions = func() []*discordgo.ApplicationCommandInteractionDataOption {
					return opt
				}

				for _, o := range opt {
					if o.Focused {
						index := slices.IndexFunc(cmd.Options, func (option *Option) bool {return option.Name == o.Name})
						if index < 0 {
							continue
						}

						option := cmd.Options[index]
						if option.Autocomplete != nil {
							(*option.Autocomplete)(&ctx, session, event)
						}
					}
				}
			}
		}
	}
}

func ConvertOptions(options []*Option) []*discordgo.ApplicationCommandOption {
	var converted []*discordgo.ApplicationCommandOption = make([]*discordgo.ApplicationCommandOption, len(options))
	for _, opt := range options {
		conv := &discordgo.ApplicationCommandOption{
			Name: opt.Name,
			Description: opt.Description,
			Required: opt.Required,
			Autocomplete: opt.Autocomplete != nil,
			Choices: opt.Choices,
		}

		if opt.Options != nil {
			conv.Options = ConvertOptions(opt.Options)
			continue
		}

		switch opt.Type.(type) {
		case string:
			conv.Type = discordgo.ApplicationCommandOptionString

		case bool:
			conv.Type = discordgo.ApplicationCommandOptionBoolean

		case int, uint:
			conv.Type = discordgo.ApplicationCommandOptionInteger

		case float32, float64:
			conv.Type = discordgo.ApplicationCommandOptionNumber

		case discordgo.Channel, *discordgo.Channel:
			conv.Type = discordgo.ApplicationCommandOptionChannel

		case discordgo.User, *discordgo.User, discordgo.Member, *discordgo.Member:
			conv.Type = discordgo.ApplicationCommandOptionUser

		case discordgo.Role, *discordgo.Role:
			conv.Type = discordgo.ApplicationCommandOptionRole

		case discordgo.MessageAttachment, *discordgo.MessageAttachment:
			conv.Type = discordgo.ApplicationCommandOptionAttachment
		}

		converted = append(converted, conv)
	}

	return converted
}

// type safety blasphemy

type WithOptions interface {
	AppendOption(opt *discordgo.ApplicationCommandOption)
}

type CommandWrapper struct {
	*discordgo.ApplicationCommand
}

func (cw *CommandWrapper) AppendOption(opt *discordgo.ApplicationCommandOption) {
	cw.Options = append(cw.Options, opt)
}

type OptionWrapper struct {
	*discordgo.ApplicationCommandOption
}

func (ow *OptionWrapper) AppendOption(opt *discordgo.ApplicationCommandOption) {
	ow.Options = append(ow.Options, opt)
}

func compact(command Command, target WithOptions) {
	for _, sub := range command.Subcommands {
		opt := &discordgo.ApplicationCommandOption{
			Type: discordgo.ApplicationCommandOptionSubCommand,
			Name: sub.Name,
			Description: sub.Description,
			Options: ConvertOptions(sub.Options),
		}

		if len(sub.Subcommands) > 0 {
			opt.Type = discordgo.ApplicationCommandOptionSubCommandGroup
			compact(*sub, &OptionWrapper{opt})
		}

		target.AppendOption(opt)
	}
}

func TraverseCommand(command Command, options []*discordgo.ApplicationCommandInteractionDataOption) (Command, []*discordgo.ApplicationCommandInteractionDataOption) {
	var collected []*discordgo.ApplicationCommandInteractionDataOption
	for _, option := range options {
		switch option.Type {
		case discordgo.ApplicationCommandOptionSubCommandGroup, discordgo.ApplicationCommandOptionSubCommand:
			o_ind := slices.IndexFunc(command.Subcommands, func (s *Command) bool {return s.Name == option.Name})
			if o_ind > -1 {
				return TraverseCommand(*command.Subcommands[o_ind], option.Options)
			} else {
				return TraverseCommand(command, option.Options)
			}

		default:
			collected = append(collected, option)
		}
	}

	return command, collected
} 