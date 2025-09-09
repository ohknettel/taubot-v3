package handlers

import (
	"github.com/bwmarrin/discordgo"
	"slices"
	"os"
)

func MigrateCommands(session *discordgo.Session, commands []Command) (map[string]Command, error) {
	var dict map[string]Command = make(map[string]Command, len(commands))
	for _, cmd := range commands {
		command := &discordgo.ApplicationCommand{Name: cmd.Name, Description: cmd.Description, DefaultMemberPermissions: cmd.DefaultPermissions, Options: cmd.Options}
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
				cmd, opt := find_cmd_from_options(h, options)
				ctx.GetOptions = func() []*discordgo.ApplicationCommandInteractionDataOption {
					return opt
				}

				cmd.Callback(&ctx, session, event)
			}

		case discordgo.InteractionApplicationCommandAutocomplete:
			options := event.ApplicationCommandData().Options
			if h, ok := commands[event.ApplicationCommandData().Name]; ok {
				cmd, opt := find_cmd_from_options(h, options)

				if cmd.Autocomplete == nil {
					return
				}

				ctx.GetOptions = func() []*discordgo.ApplicationCommandInteractionDataOption {
					return opt
				}

				for _, o := range opt {
					if _, ok = (*cmd.Autocomplete)[o.Name]; ok && o.Focused {
						(*cmd.Autocomplete)[o.Name](&ctx, session, event)
					}
				}
			}
		}
	}
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
	// going through indexes is less expensive than shallow copying
	for i := 0; i < len(command.Subcommands); i++ {
		sub := command.Subcommands[i]
		opt := &discordgo.ApplicationCommandOption{
			Type: discordgo.ApplicationCommandOptionSubCommand,
			Name: sub.Name,
			Description: sub.Description,
			Options: sub.Options,
		}

		if len(sub.Subcommands) > 0 {
			opt.Type = discordgo.ApplicationCommandOptionSubCommandGroup
			compact(*sub, &OptionWrapper{opt})
		}

		target.AppendOption(opt)
	}
}

func find_cmd_from_options(command Command, options []*discordgo.ApplicationCommandInteractionDataOption) (Command, []*discordgo.ApplicationCommandInteractionDataOption) {
	var collected []*discordgo.ApplicationCommandInteractionDataOption
	for _, option := range options {
		switch option.Type {
		case discordgo.ApplicationCommandOptionSubCommandGroup, discordgo.ApplicationCommandOptionSubCommand:
			o_ind := slices.IndexFunc(command.Subcommands, func (s *Command) bool {return s.Name == option.Name})
			if o_ind > -1 {
				return find_cmd_from_options(*command.Subcommands[o_ind], option.Options)
			} else {
				return find_cmd_from_options(command, option.Options)
			}

		default:
			collected = append(collected, option)
		}
	}

	return command, collected
} 