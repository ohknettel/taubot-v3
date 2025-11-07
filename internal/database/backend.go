package database

import (
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

var RecordNotFoundError error = gorm.ErrRecordNotFound 

type Backend struct {
	db *gorm.DB
}

type SessionedMember struct {
	discordgo.Member
	Session *discordgo.Session
}

type BackendError struct {
	error
	Message string
}

func (err BackendError) Error() string {
	return err.Message
}

func NewBackend(database *gorm.DB) *Backend {
	return &Backend{
		db: database,
	}
}

func EvaluatePrecedence(permission UserPermission) uint8 {
	if permission.AccountID == nil && permission.EconomyID == nil {
		return 1
	} else if permission.AccountID == nil {
		return 2
	}
	return 3
}

func (self *Backend) GetPermissions(member discordgo.Member) ([]UserPermission, error) {
	var permissions []UserPermission

	stmt := self.db.Where("user_id = ?", member.User.ID)
	for _, role := range member.Roles {
		stmt = stmt.Or("user_id = ?", role)
	}

	if err := stmt.Find(&permissions).Error; err != nil {
		return []UserPermission{}, err
	}

	return permissions, nil
}

func (self *Backend) HasPermission(member SessionedMember, permission uint8, account *Account, economy *Economy) (bool, error) {
	var permissions []UserPermission
	var err error

	stmt := self.db.Where("permission_id = ?", permission).Where("user_id IN ?", append(member.Roles, member.User.ID))

	if account != nil {
		stmt = stmt.Where("account_id = ?", account.ID)
	}

	if economy != nil {
		stmt = stmt.Where("economy_id = ?", account.ID)
	}

	err = stmt.Find(&permissions).Error
	if err != nil {
		return false, err
	} else if len(permissions) == 0 {
		return false, nil
	}

	best := permissions[0]
	for _, perm := range permissions {
		if EvaluatePrecedence(best) > EvaluatePrecedence(perm) {
			best = perm
			continue
		} else if EvaluatePrecedence(best) == EvaluatePrecedence(perm) {
			if best.UserID == member.User.ID {
				continue
			} else if perm.UserID == member.User.ID {
				best = perm
			} else if member.Session != nil {
				role_best, err_best := member.Session.State.Role(member.GuildID, best.UserID)
				role_perm, err_perm := member.Session.State.Role(member.GuildID, perm.UserID)
				if err_best != nil || err_perm != nil || role_best == nil || role_perm == nil {
					continue
				} else if role_best.Position < role_perm.Position {
					best = perm
				}
			}
		}
	}

	return best.Value, nil
}

func (self *Backend) GetEconomies() ([]Economy, error) {
	var economies []Economy
	result := self.db.Find(&economies)
	if err := result.Error; err != nil {
		return nil, err
	}
	return economies, nil
}

func (self *Backend) GetEconomiesIn(guild_id string) ([]Economy, error) {
	var economies []Economy
	result := self.db.Joins("JOIN guilds ON guilds.economy_id = economies.id").Where("guilds.guild_id = ?", guild_id).Preload("Guilds").Find(&economies)
	if err := result.Error; err != nil {
		return nil, err
	}
	return economies, nil
}

func (self *Backend) GetEconomyByID(id string) (Economy, error) {
	var economy Economy
	result := self.db.Where("id = ?", id).First(&economy)
	if err := result.Error; err != nil {
		return Economy{}, err
	}
	return economy, nil
}

func (self *Backend) GetEconomyByName(name string) (Economy, error) {
	var economy Economy
	result := self.db.Where("name = ?", name).First(&economy)
	if err := result.Error; err != nil {
		return Economy{}, err
	}
	return economy, nil
}

func (self *Backend) RegisterGuild(member SessionedMember, guild discordgo.Guild, economy Economy) error {
	if perm, err := self.HasPermission(member, P_ManageEconomies, nil, &economy); err == nil && !perm {
		return BackendError{Message: "You do not have the permission to register guilds to this economy."}
	}

	session := self.db.Begin()

	var g Guild	
	guild_d := session.Where("guild_id = ?", guild.ID).First(&g)
	if err := guild_d.Error; err == nil {
		if err := self.UnregisterGuildTx(session, guild, economy); err != nil {
			session.Rollback()
			return err
		}
	}

	err := session.Model(&economy).Association("Guilds").Append(&Guild{
		GuildID: guild.ID,
	})
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit().Error
}

func (self *Backend) UnregisterGuild(member SessionedMember, guild discordgo.Guild, economy Economy) error {
	if perm, err := self.HasPermission(member, P_ManageEconomies, nil, &economy); err == nil && !perm {
		return BackendError{Message: "You do not have the permission to unregister guilds from this economy."}
	}

	session := self.db.Begin()

	var g Guild	
	guild_d := session.Where("guild_id = ?", guild.ID).First(&g)
	if err := guild_d.Error; err != nil {
		return err
	}

	err := session.Model(&economy).Association("Guilds").Delete(&Guild{GuildID: guild.ID})
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit().Error
}

func (self *Backend) UnregisterGuildTx(session *gorm.DB, guild discordgo.Guild, economy Economy) error {
	var g Guild	
	guild_d := session.Where("guild_id = ?", guild.ID).First(&g)
	if err := guild_d.Error; err != nil {
		return err
	}

	err := session.Model(&economy).Association("Guilds").Delete(&Guild{GuildID: guild.ID})
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit().Error
}

func (self *Backend) CreateEconomy(member SessionedMember, economy *Economy) error {
	if perm, err := self.HasPermission(member, P_ManageEconomies, nil, nil); err == nil && !perm {
		return BackendError{Message: "You do not have the permission to create economies."}
	}

	session := self.db.Begin()

	err := session.Where("name = ?", economy.Name).FirstOrCreate(economy).Error
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit().Error
}

func (self *Backend) DeleteEconomy(member SessionedMember, economy *Economy) error {
	if perm, err := self.HasPermission(member, P_ManageEconomies, nil, nil); err == nil && !perm {
		return BackendError{Message: "You do not have the permission to delete economies."}
	}

	session := self.db.Begin()

	err := session.Delete(economy).Error
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit().Error
}
