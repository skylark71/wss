package logic

import (
	"encoding/json"
	"single-win-system/configs"
	mdels "single-win-system/models"

	log "github.com/sirupsen/logrus"
)

const (
	USER_BLOCKED = 20
	USER_DELETED = 30

	TYPE_SLOT_BET               = 1
	TYPE_RECEIVING_GB_FROM_USER = 7
	TYPE_AR_RELOAD              = 12
)

func incWinVal(prize float32) float32 {
	var game_tokens_configuration mdels.Game_tokens_configuration
	_, err := configs.Db.QueryOne(&game_tokens_configuration, `SELECT * FROM game_tokens_configuration WHERE id=?`, 1)
	if err != nil {
		log.Warn("[incWinVal] Error Select game_tokens_configuration %v", err.Error())
	}

	log.Infof("[incWinVal] game_tokens_configuration %v ", game_tokens_configuration)

	prize = prize + (prize * float32(game_tokens_configuration.Ar_incrementation_percent) / 100)

	return prize
}

func isDailyBonusReceived(user mdels.User) bool {
	var bonus_history mdels.Bonus_history
	_, err := configs.Db.QueryOne(&bonus_history, `SELECT * FROM bonus_history WHERE date=? AND user_id=?`, GetEkbTime().Format("2006-01-02"), user.Id)
	log.Infof("[BonusMachineSpinRequest] bonus_history %v", bonus_history)
	if err != nil {
		log.Warn("[isDailyBonusReceived] Error SELECT ", err.Error())
		return false
	} else {
		return true
	}
}

func isThatNotProductOrTransferExpenses(type_id int) bool {
	if type_id == TYPE_SLOT_BET || type_id == TYPE_AR_RELOAD {
		return true
	} else {
		return false
	}
}

func fightWithObscure(fight mdels.Fight, user_id int) mdels.Fight {
	if fight.Obscure.Armour > 0 {
		fight.Obscure.Armour = fight.Obscure.Armour - 1
		fight.Obscure.Battery = fight.Obscure.Battery - fight.Battery_armour
		val := -float32(fight.Battery_armour)
		UpdateGameHistory(user_id, 7, &val, nil)
		return fight
	}

	if fight.Obscure.Armour <= 0 {
		fight.Obscure.Hp = fight.Obscure.Hp - 1
		fight.Obscure.Battery = fight.Obscure.Battery - fight.Battery_life
		val := -float32(fight.Battery_life)
		UpdateGameHistory(user_id, 7, &val, nil)
		return fight
	}

	return fight
}

func GetObscure(user_id int) mdels.Fight {
	var cfg_game mdels.Cfg_game
	_, err := configs.Db.QueryOne(&cfg_game, `SELECT * FROM cfg_game WHERE is_active = ?`, true)
	if err != nil && err.Error() != "pg: multiple rows in result set" {
		log.Warn("[getObscure] Error SELECT ", err.Error())
	}
	log.Infof("[getObscure] cfg_game %v", cfg_game)

	var settings_obscure mdels.Settings_obscure
	var array_obj []string
	var battery_armour, battery_life, obscure_charge, obscure_storage int
	var obscure_life float32
	array_obj = append(array_obj, cfg_game.Battery_armour, cfg_game.Battery_life, cfg_game.Obscure_charge, cfg_game.Obscure_life, cfg_game.Obscure_storage)
	for i := 0; i < len(array_obj); i++ {
		err = json.Unmarshal([]byte(array_obj[i]), &settings_obscure)
		switch i {
		case 0:
			battery_armour = randInt(settings_obscure.Down, settings_obscure.Up)
		case 1:
			battery_life = randInt(settings_obscure.Down, settings_obscure.Up)
		case 2:
			obscure_charge = randInt(settings_obscure.Down, settings_obscure.Up)
		case 3:
			obscure_life = float32(randInt(settings_obscure.Down, settings_obscure.Up))
		case 4:
			obscure_storage = randInt(settings_obscure.Down, settings_obscure.Up)
		}
	}
	log.Infof("[getObscure] obscure_life %v", obscure_life)
	log.Infof("[getObscure] battery_armour %v", battery_armour)
	log.Infof("[getObscure] battery_life %v", battery_life)
	log.Infof("[getObscure] obscure_charge %v", obscure_charge)
	log.Infof("[getObscure] obscure_storage %v", obscure_storage)

	var obscure mdels.Obscure
	obscure.Armour = randInt(int(obscure_life*0.1), int(obscure_life*0.7))
	obscure.Hp = int(obscure_life) - obscure.Armour
	battery, err := GetBattery(cfg_game.Battery, user_id)
	if err != nil {
		log.Warn("[getObscure] getBattery %v", err.Error())
	}
	obscure.Battery = int(battery)
	log.Infof("[getObscure] Obscure %v", obscure)

	var fight mdels.Fight
	fight.Obscure = obscure
	fight.Battery_armour = battery_armour
	fight.Battery_life = battery_life
	fight.Obscure_charge = obscure_charge
	fight.Obscure_storage = obscure_storage

	// date := GetEkbTime()
	// date_bonus := GetBonusDate(user_id)
	// log.Infof("date: %v", date)
	// log.Infof("date_bonus: %v", date_bonus.Format(time.RFC3339))
	// if date_bonus.Format(time.RFC3339) > date.Format(time.RFC3339) {
	// 	fight.Battery_armour = 0
	// 	fight.Battery_life = 0
	// 	UpdateBattery(100, user_id)
	// }
	return fight
}
