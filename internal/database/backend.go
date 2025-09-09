package database

import (
	"slices"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

var RecordNotFoundError error = gorm.ErrRecordNotFound 

type Backend struct {
	db *gorm.DB
}

func NewBackend(database *gorm.DB) *Backend {
	return &Backend{
		db: database,
	}
}

func (self *Backend) GetPermissions(member discordgo.Member) ([]UserPermission, error) {
	var permissions []UserPermission

	result := self.db.Where("user_id = ?", member.User.ID).Find(&permissions)
	if err := result.Error; err != nil {
		return nil, err
	}
	
	return permissions, nil
}

func (self *Backend) HasPermissionsAny(member discordgo.Member, perms []uint8, model any) (bool, error) {
	var permissions []UserPermission
	var err error

	switch model := model.(type) {
	case Economy:
		err = self.db.Where("user_id = ?", member.User.ID).Where("economy_id IS NOT NULL AND economy_id = ?", model.ID).Find(&permissions).Error

	case Account:
		err = self.db.Where("user_id = ?", member.User.ID).Where("account_id IS NOT NULL AND account_id = ?", model.ID).Find(&permissions).Error

	default:
		err = self.db.Where("user_id = ?", member.User.ID).Where("account_id IS NULL AND economy_id IS NULL").Find(&permissions).Error
	}

	if err != nil || len(permissions) == 0 {
		return false, err
	}

	for _, permission := range permissions {
		if slices.Contains(perms, permission.PermissionID) {
			return true, nil
		}
	}

	return false, nil
}

func (self *Backend) HasPermissionsAll(member discordgo.Member, perms []uint8, model any) (bool, error) {
	var permissions []UserPermission
	var err error

	switch model := model.(type) {
	case Economy:
		err = self.db.Where("user_id = ?", member.User.ID).Where("economy_id IS NOT NULL AND economy_id = ?", model.ID).Find(&permissions).Error

	case Account:
		err = self.db.Where("user_id = ?", member.User.ID).Where("account_id IS NOT NULL AND account_id = ?", model.ID).Find(&permissions).Error

	default:
		err = self.db.Where("user_id = ?", member.User.ID).Where("account_id IS NULL AND economy_id IS NULL").Find(&permissions).Error
	}

	if err != nil || len(permissions) == 0 {
		return false, err
	}

	for _, permission := range permissions {
		if !slices.Contains(perms, permission.PermissionID) {
			return false, nil
		}
	}

	return true, nil
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

func (self *Backend) RegisterGuild(guild discordgo.Guild, economy Economy) error {
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

func (self *Backend) UnregisterGuild(guild discordgo.Guild, economy Economy) error {
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

func (self *Backend) CreateEconomy(economy *Economy) error {
	session := self.db.Begin()

	err := session.Where("name = ?", economy.Name).FirstOrCreate(economy).Error
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit().Error
}

func (self *Backend) DeleteEconomy(economy *Economy) error {
	session := self.db.Begin()

	err := session.Delete(economy).Error
	if err != nil {
		session.Rollback()
		return err
	}

	return session.Commit().Error
}
