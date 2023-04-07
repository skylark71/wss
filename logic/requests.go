package logic

import (
	"errors"
	"math"
	"single-win-system/configs"
	mdels "single-win-system/models"

	log "github.com/sirupsen/logrus"
)

const (
	BONUS_BET = 20
	AR_BET    = 40
)

func ArShootReq(count int, fight mdels.Fight, user mdels.User, place_id string, place_location int, user_location int, place_type int, game_token int) (mdels.Fight, mdels.Data, error) {
	if count <= 0 {
		if len(place_id) == 0 {
			return mdels.Fight{}, mdels.Data{}, errors.New("Точка не найдена")
		}
		ar_shoot_history, _ := getArShootHistory(place_id, place_type, user.Id)
		log.Info("[ArShootReq] ar_shoot_history", ar_shoot_history)
		if ar_shoot_history == (mdels.Ar_shoot_history{}) {
			ar_shoot_history.Place_id = place_id
			ar_shoot_history.Place_type = place_type
			ar_shoot_history.User_id = user.Id
			ar_shoot_history.Created_at = GetEkbTime()
			log.Info("[ArShootReq] ar_shoot_history ", ar_shoot_history)
			insertArShootHistory(ar_shoot_history)
		}

		if game_token != 0 {
			user.Token_balance -= 1
			updateUserTokenBalance(user)
		}

		updateArShootHistoryTry(ar_shoot_history.Try+1, user, place_id, place_type)

		user.Shots = user.Shots - 1
		updateUserShoots(user)
	}

	flag, _ := CheckInfBatter(user.Id)
	if flag {
		fight.Battery_armour = 0.0
		fight.Battery_life = 0.0
		UpdateBattery(100, user.Id)
		fight.Obscure.Battery = 100
	}

	fight = fightWithObscure(fight, user.Id)
	if fight.Obscure.Battery <= 0 {
		UpdateBattery(0, user.Id)
		return mdels.Fight{}, mdels.Data{}, errors.New("Проигрыш")
	} else {
		UpdateBattery(float32(fight.Obscure.Battery), user.Id)
	}

	if fight.Obscure.Hp <= 0 && fight.Obscure.Armour <= 0 {
		ar_shoot_history, _ := getArShootHistory(place_id, place_type, user.Id)

		log.Infof("[ArShootReq] WIN user  %v", user)
		log.Infof("[ArShootReq] WIN ar_shoot_history  %v", ar_shoot_history)

		isWin, prize, _ := Spin(AR_BET, AR_MACHINE, user.Id)
		log.Infof("[ArShootReq] Win %v, Prize: %v", isWin, prize)

		var data mdels.Data
		data.Blockdata = fight.Obscure_storage
		data.Battery = fight.Obscure_charge
		var val float32
		if flag {
			val = 100
		} else {
			val = float32(fight.Obscure.Battery + fight.Obscure_charge)
		}
		UpdateGameHistory(user.Id, 6, &val, &data.Blockdata)
		UpdateBattery(val, user.Id)
		data.Blockdata = UpdateStorage(fight.Obscure_storage, user.Id)

		UpdateArShootHistoryPrize(prize, user, place_id, place_type, fight.Obscure_storage, fight.Obscure_charge)
		if isWin {
			data.Prize = prize
			if game_token != 0 {
				prize = incWinVal(prize)
				prize = float32(math.Round(float64(prize)))
			}
			UpdateBalance(user, prize, 11, &ar_shoot_history.Id)
		}

		return fight, data, nil
	}

	return fight, mdels.Data{}, nil
}

func SlotMachineSpinRequest(bet int, demo bool, user mdels.User) (*bool, *float32, *[]int, mdels.User, error) {
	var slot_configuration mdels.Slot_configuration
	_, err := configs.Db.QueryOne(&slot_configuration, `SELECT * FROM slot_configuration`)
	if err != nil {
		log.Warn("[UpdateBalance] Error Select slot_configuration", err.Error())
	}
	log.Infof("slot_configuration", slot_configuration)

	if bet < slot_configuration.Min_bet || bet > slot_configuration.Max_bet {
		log.Warn("Incorrect bet value.")
		return nil, nil, nil, user, err
	}

	if !demo {
		if user.Balance < bet {
			log.Warn("Not enough money on balance to place a bet.")
			return nil, nil, nil, user, err
		}
		UpdateBalance(user, float32(-bet), 1, nil)
		user.Balance = user.Balance - bet
	}

	var history mdels.Slot_history
	isWin, prize, combination := Spin(bet, SLOT_MACHINE, user.Id)
	if !demo {
		tmpPrize := float32(math.Ceil(float64(prize)))
		prizeInt := int(tmpPrize)
		combinationS := ArrayToString(combination, ",")
		history = AddSlotHistory(user.Id, bet, &prizeInt, combinationS)
	}

	if isWin && !demo {
		UpdateBalance(user, prize, 2, &history.Id)
		user.Balance = user.Balance + int(prize)
	}

	log.Info(user)

	return &isWin, &prize, &combination, user, nil
}

func BonusMachineSpinRequest(token string) *float32 {
	var user mdels.User
	_, err := configs.Db.QueryOne(&user, `SELECT "user".* FROM "user" JOIN user_tokens ON user_tokens.user_id = "user".id WHERE user_tokens.access_token = ?`, token)
	if err != nil {
		log.Warn("[ArShootReq] Error SELECT ", err.Error())
	}
	if isDailyBonusReceived(user) {
		log.Info("[BonusMachineSpinRequest] Daily bonus already received.")
		return nil
	}

	var bonus_events mdels.Bonus_events
	_, err = configs.Db.QueryOne(&bonus_events, `SELECT * FROM bonus_events WHERE type_id=?`, 1)
	if err != nil {
		log.Warn("[BonusMachineSpinRequest] Error SELECT bonus_events", err.Error())
	}
	log.Infof("[BonusMachineSpinRequest] bonus_events", bonus_events)

	prize := SpinBonus(BONUS_BET, BONUS_MACHINE, user.Id)
	log.Infof("[BonusMachineSpinRequest] prize", prize)
	history := AddBonusHistory(user.Id, int(prize))
	UpdateBalance(user, prize, 13, &history.Id)

	return &prize
}
