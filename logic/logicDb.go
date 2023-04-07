package logic

import (
	"encoding/json"
	"math/rand"
	"single-win-system/configs"
	mdels "single-win-system/models"
	"time"

	log "github.com/sirupsen/logrus"
)

func UpdateBalance(user mdels.User, amount float32, type_id int, linked_history_id *int) {
	log.Infof("[UpdateBalance] [%v] old balance %v", user.Id, user.Balance)
	new_balance := float32(user.Balance) + amount
	_, err := configs.Db.QueryOne(&user, `UPDATE "user" SET balance=? WHERE id=?`, new_balance, user.Id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[UpdateBalance] [%v] Error UPDATE ", user.Id, err.Error())
	}
	log.Infof("[UpdateBalance] [%v] new balance %v", user.Id, new_balance)
	addBalanceOperation(user.Id, amount, type_id, linked_history_id, new_balance)

	if amount > 0 && type_id != TYPE_RECEIVING_GB_FROM_USER || isThatNotProductOrTransferExpenses(type_id) {
		updateMontlyWinHistory(user.Id, int(amount))
	}
}

func addBalanceOperation(user_id int, amount float32, type_id int, linked_history_id *int, balance float32) {
	date := GetEkbTime()
	var balance_history mdels.Balance_history
	balance_history.Amount = amount
	balance_history.Balance = balance
	balance_history.User_id = user_id
	balance_history.Date = date
	balance_history.Type_id = type_id
	balance_history.Linked_history_id = linked_history_id
	_, err := configs.Db.QueryOne(&balance_history, `INSERT INTO balance_history (amount, balance, user_id, date, type_id, linked_history_id) VALUES (?amount, ?balance, ?user_id, ?date, ?type_id, ?linked_history_id)RETURNING id`, balance_history)
	if err != nil {
		log.Warn("[addBalanceOperation] [%v] Insert Error ", user_id, err.Error())
	}
	log.Infof("[addBalanceOperation] [%v] balance_history %v", user_id, balance_history)
}

func updateMontlyWinHistory(user_id int, amount int) {
	first_day_current_month := getFirstDayOfMonth()
	var user_monthly_win_history mdels.User_monthly_win_history
	_, err := configs.Db.QueryOne(&user_monthly_win_history, `SELECT * FROM user_monthly_win_history WHERE user_id = ? AND date=?`, user_id, first_day_current_month)
	log.Infof("[updateMontlyWinHistory] [%v] user_monthly_win_history %v", user_id, user_monthly_win_history)
	if err != nil {
		if err.Error() == "pg: no rows in result set" {
			rand.Seed(time.Now().UTC().UnixNano())
			user_monthly_win_history.User_id = user_id
			user_monthly_win_history.Date = first_day_current_month
			user_monthly_win_history.Win_limit = randInt(27000, 35000)
			user_monthly_win_history.Amount = amount
			log.Infof("[updateMontlyWinHistory] [%v] user_monthly_win_history %v", user_id, user_monthly_win_history)
			_, err := configs.Db.QueryOne(&user_monthly_win_history, `INSERT INTO user_monthly_win_history (user_id, date, win_limit, amount) VALUES (?user_id, ?date, ?win_limit, ?amount)RETURNING id`, user_monthly_win_history)
			if err != nil {
				log.Warn("[updateMontlyWinHistory] [%v] Insert Error ", user_id, err.Error())
			}
		}
		log.Warn("[updateMontlyWinHistory] [%v] Select Error ", user_id, err.Error())
		return
	}
	_, err = configs.Db.QueryOne(&user_monthly_win_history, `UPDATE user_monthly_win_history SET amount=? WHERE id=?`, user_monthly_win_history.Amount+amount, user_monthly_win_history.Id)
	log.Infof("[updateMontlyWinHistory] [%v] user_monthly_win_history %v", user_id, user_monthly_win_history.Amount+amount)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[updateMontlyWinHistory] [%v] Error UPDATE ", user_id, err.Error())
	}
}

func AddSlotHistory(user_id int, bet int, prize *int, combination string) mdels.Slot_history {
	var slot_history mdels.Slot_history
	slot_history.User_id = user_id
	slot_history.Bet = bet
	slot_history.Prize = prize
	slot_history.Combination = combination
	slot_history.Created_at = GetEkbTime()
	_, err := configs.Db.QueryOne(&slot_history, `INSERT INTO slot_history (user_id, bet, prize, combination, created_at) VALUES (?user_id, ?bet, ?prize, ?combination, ?created_at)RETURNING id`, slot_history)
	if err != nil {
		log.Warn("[addSlotHistory] [%v] Insert Error ", user_id, err.Error())
	}

	return slot_history
}

func AddBonusHistory(user_id int, amount int) mdels.Bonus_history {
	var bonus_history mdels.Bonus_history
	bonus_history.User_id = user_id
	bonus_history.Date = GetEkbTime()
	bonus_history.Datetime = GetEkbTime()
	bonus_history.Amount = amount
	_, err := configs.Db.QueryOne(&bonus_history, `INSERT INTO bonus_history (user_id, date, datetime, amount) VALUES (?user_id, ?date, ?datetime, ?amount)RETURNING id`, bonus_history)
	if err != nil {
		log.Warn("[AddBonusHistory] [%v] Insert Error bonus_history", user_id, err.Error())
	}

	return bonus_history
}

func GetBattery(def_obscure_charge int, user_id int) (float32, error) {
	var user_battery mdels.User_battery
	_, err := configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[getBattery] [%v] Error SELECT user_battery", user_id, err.Error())
		user_battery.Battery = float32(def_obscure_charge)
		user_battery.User_id = user_id
		_, err = configs.Db.QueryOne(&user_battery, `INSERT INTO user_battery (user_id, battery) VALUES (?user_id, ?battery)RETURNING id`, user_battery)
		if err != nil {
			log.Warn("[getBattery] [%v] Error INSERT user_battery", user_id, err.Error())
		}

		return user_battery.Battery, err
	}
	log.Infof("[getBattery] [%v] user_battery %v", user_id, user_battery)
	return user_battery.Battery, nil
}

func GetMonthlyWinHistory(user_id int) mdels.User_monthly_win_history {
	var user_monthly_win_history mdels.User_monthly_win_history
	first_day_current_month := getFirstDayOfMonth()
	_, err := configs.Db.QueryOne(&user_monthly_win_history, `SELECT * FROM user_monthly_win_history WHERE user_id = ? and date >= ?`, user_id, first_day_current_month)
	if err != nil {
		log.Warn("[GetMonthlyWinHistory] [%v] Error SELECT user_monthly_win_history", user_id, err.Error())
		if err.Error() == "pg: no rows in result set" {
			updateMontlyWinHistory(user_id, 0)
		}
		log.Warn("[updateMontlyWinHistory] [%v] Select Error ", user_id, err.Error())
	}
	log.Infof("[GetMonthlyWinHistory] [%v] user_monthly_win_history %v", user_id, user_monthly_win_history)
	return user_monthly_win_history
}

func GetUpdateBattery(user_id int) (float32, error) {
	var user_battery mdels.User_battery
	_, err := configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[GetUpdateBattery] [%v] Error SELECT user_battery", user_id, err.Error())
		return user_battery.Battery, err
	}

	var user mdels.User
	_, err = configs.Db.QueryOne(&user, `SELECT * FROM "user" WHERE id = ?`, user_id)
	if err != nil {
		log.Warn("[GetUpdateBattery] [%v] Error SELECT user_battery", user_id, err.Error())
		return user_battery.Battery, err
	}
	timein := time.Now()
	diff := timein.Sub(user.Last_online)
	v := int(diff.Seconds())
	battery, err := AddTime(v, user_id)
	log.Infof("[GetUpdateBattery] [%v] CurrentTime: %v LastOnline %v Diff %v", user_id, timein, user.Last_online, v)
	_, err = configs.Db.QueryOne(&user, `UPDATE "user" SET last_online=? WHERE id = ?`, timein, user_id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[GetUpdateBattery] [%v] Error UPDATE user", user_id, err.Error())
	}

	if battery >= 100 {
		UpdateBattery(100, user_id)
		return 100, nil
	}

	log.Infof("[GetUpdateBattery] [%v] user_battery %v", user_id, user_battery.Battery)
	return battery, nil
}

func GetInfBattery(user_id int) (int, error) {
	var user_battery mdels.User_battery
	_, err := configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[GetInfBattery] [%v] Error SELECT user_battery", user_id, err.Error())
		return 0, err
	}

	currentTime := GetEkbTime()
	loc := time.FixedZone("UTC+5", 5*60*60)
	t := user_battery.Bonus_end.In(loc)
	tt := t.Add(-time.Hour * 5)
	diff := tt.Sub(currentTime)
	log.Infof("[GetInfBattery] [%v] CurrentTime: %v BonusEnd %v Diff %v", user_id, currentTime, tt, diff.Minutes())
	return int(diff.Minutes()), nil
}

func CheckInfBatter(user_id int) (bool, error) {
	var user_battery mdels.User_battery
	_, err := configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[GetInfBattery] [%v] Error SELECT user_battery", user_id, err.Error())
		return false, err
	}

	if user_battery.Bonus_start.IsZero() || user_battery.Bonus_end.IsZero() {
		return false, nil
	}

	currentTime := GetEkbTime()
	loc := time.FixedZone("UTC+5", 5*60*60)
	t := user_battery.Bonus_end.In(loc)
	tt := t.Add(-time.Hour * 5)
	diff := tt.Sub(currentTime)
	log.Infof("[GetInfBattery] [%v] CurrentTime: %v BonusEnd %v Diff %v", user_id, currentTime, tt, diff.Minutes())

	if int(diff.Minutes()) > 0 {
		return true, nil
	}

	return false, nil
}

func SetTutorialAr(user mdels.User) {
	_, err := configs.Db.QueryOne(&user, `UPDATE "user" SET is_tutorial_finished=? WHERE id=?`, 1, user.Id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[SetTutorialAr] [%v] Error UPDATE ", user.Id, err.Error())
	}
}

func GetBonusDate(user_id int) time.Time {
	var user_battery mdels.User_battery
	_, err := configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[GetBonusDate] [%v] Error SELECT user_battery", user_id, err.Error())
		return time.Time{}
	}
	log.Infof("[GetBonusDate] [%v] get_bonus_date %v", user_id, user_battery)
	return user_battery.Bonus_end
}

func UpdateBattery(battery float32, user_id int) {
	if battery >= 100 {
		battery = 100
	}
	var user_battery mdels.User_battery
	_, err := configs.Db.QueryOne(&user_battery, `UPDATE user_battery SET battery=? WHERE user_id = ?`, battery, user_id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[updateBattery] [%v] Error UPDATE ar_shot_history", user_id, err.Error())
	}
	log.Infof("[UpdateBattery] [%v] user_battery %v", user_id, battery)
}

func UpdateStorage(storage int, user_id int) int {
	var user_battery mdels.User_battery
	_, err := configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[UpdateStorage] [%v] Error SELECT user_battery", user_id, err.Error())
	}

	user_battery.Block_data = user_battery.Block_data + float32(storage)
	_, err = configs.Db.QueryOne(&user_battery, `UPDATE user_battery SET block_data=? WHERE user_id = ?`, user_battery.Block_data, user_id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[UpdateStorage] [%v] Error UPDATE ar_shot_history", user_id, err.Error())
	}
	log.Infof("[UpdateStorage] [%v] storage %v", user_id, storage)
	return storage
}

func getArShootHistory(place_id string, place_type, user_id int) (mdels.Ar_shoot_history, error) {
	var ar_shoot_history mdels.Ar_shoot_history
	_, err := configs.Db.QueryOne(&ar_shoot_history, `SELECT * FROM ar_shot_history WHERE place_id = ? AND place_type=? AND user_id =?`, place_id, place_type, user_id)
	if err != nil {
		log.Warn("[getArShootHistory] [%v] Error SELECT ", user_id, err.Error())
		return ar_shoot_history, err
	}
	return ar_shoot_history, nil
}

func insertArShootHistory(ar_shoot_history mdels.Ar_shoot_history) {
	_, err := configs.Db.QueryOne(&ar_shoot_history, `INSERT INTO ar_shot_history (place_id, place_type, user_id, created_at) VALUES (?place_id, ?place_type, ?user_id, ?created_at)RETURNING id`, ar_shoot_history)
	if err != nil {
		log.Warn("[insertArShootHistory] [%v] Error INSERT ", ar_shoot_history.User_id, err.Error())
	}
}

func updateArShootHistoryTry(try int, user mdels.User, place_id string, place_type int) {
	_, err := configs.Db.QueryOne(&user, `UPDATE ar_shot_history SET try=?, updated_at=? WHERE place_id = ? AND place_type=? AND user_id =?`, try, GetEkbTime(), place_id, place_type, user.Id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[ArShootReq] [%v] Error UPDATE ar_shot_history", user.Id, err.Error())
	}
}

func updateUserTokenBalance(user mdels.User) {
	_, err := configs.Db.QueryOne(&user, `UPDATE "user" SET token_balance=? WHERE id=?`, user.Token_balance, user.Id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[ArShootReq] [%v] Error UPDATE ", user.Id, err.Error())
	}
	log.Infof("[ArShootReq] [%v] UPDATE user by token_balance %v", user.Id, user)
}

func updateUserShoots(user mdels.User) {
	_, err := configs.Db.QueryOne(&user, `UPDATE "user" SET shots=? WHERE id=?`, user.Shots, user.Id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[ArShootReq] [%v]  Error UPDATE ", user.Id, err.Error())
	}
	log.Infof("[ArShootReq] [%v]  UPDATE user by token_balance %v", user.Id, user)
}

func UpdateArShootHistoryPrize(prize float32, user mdels.User, place_id string, place_type int, block_data int, battery int) {
	var cfg_game mdels.Cfg_game
	var settings_obscure mdels.Settings_obscure
	_, err := configs.Db.QueryOne(&cfg_game, `SELECT * FROM cfg_game WHERE is_active = ?`, true)
	if err != nil && err.Error() != "pg: multiple rows in result set" {
		log.Warn("[getObscure] Error SELECT ", err.Error())
	}
	log.Infof("[getObscure] cfg_game %v", cfg_game.Cool_down)

	if place_id == "0" {
		_, err = configs.Db.QueryOne(&user, `UPDATE ar_shot_history SET prize=?, updated_at=?, block_data=?, battery=? WHERE place_id = ? AND place_type=? AND user_id =?`, prize, GetEkbTime(), block_data, battery, place_id, place_type, user.Id)
		if err != nil && err.Error() != "pg: no rows in result set" {
			log.Warn("[ArShootReq] [%v] Error UPDATE ar_shot_history", user.Id, err.Error())
		}
		return
	}
	err = json.Unmarshal([]byte(cfg_game.Cool_down), &settings_obscure)
	value := randInt(settings_obscure.Down, settings_obscure.Up)
	t := GetEkbTime()
	t1 := t.Add(time.Minute * time.Duration(value))
	_, err = configs.Db.QueryOne(&user, `UPDATE ar_shot_history SET prize=?, updated_at=?, expired_at=?, block_data=?, battery=?  WHERE place_id = ? AND place_type=? AND user_id =?`, prize, GetEkbTime(), t1, block_data, battery, place_id, place_type, user.Id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[ArShootReq] [%v] Error UPDATE ar_shot_history", user.Id, err.Error())
	}
}

func AddSteps(steps int, user_id int) error {
	var cfg_game mdels.Cfg_game
	_, err := configs.Db.QueryOne(&cfg_game, `SELECT * FROM cfg_game WHERE is_active = ?`, true)
	if err != nil && err.Error() != "pg: multiple rows in result set" {
		log.Warn("[AddSteps] [%v] Error SELECT cfg_game", user_id, err.Error())

		return err
	}

	var user_battery mdels.User_battery
	_, err = configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[AddSteps] [%v] Error SELECT user_battery", user_id, err.Error())

		return err
	}

	steps_battery := user_battery.Battery + (float32(steps) * cfg_game.Battery_steps_charge)
	_, err = configs.Db.QueryOne(&user_battery, `UPDATE user_battery SET battery=? WHERE user_id = ?`, steps_battery, user_id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[AddSteps] [%v] Error UPDATE ar_shot_history", user_id, err.Error())

		return err
	}

	log.Infof("[AddSteps] [%v] Perc bat %f from steps %v", user_id, steps_battery, steps)

	return nil
}

func AddTime(seconds int, user_id int) (float32, error) {
	var cfg_game mdels.Cfg_game
	_, err := configs.Db.QueryOne(&cfg_game, `SELECT * FROM cfg_game WHERE is_active = ?`, true)
	if err != nil && err.Error() != "pg: multiple rows in result set" {
		log.Warn("[AddTime] [%v] Error SELECT cfg_game", user_id, err.Error())

		return 0, err
	}

	var user_battery mdels.User_battery
	_, err = configs.Db.QueryOne(&user_battery, `SELECT * FROM user_battery WHERE user_id = ?`, user_id)
	if err != nil {
		log.Warn("[AddTime] [%v] Error SELECT user_battery", user_id, err.Error())

		return 0, err
	}

	if user_battery.Battery >= 100 {
		UpdateBattery(100, user_id)
		return 100, nil
	}

	diff := float32(seconds) * cfg_game.Battery_time_charge
	steps_battery := user_battery.Battery + diff
	_, err = configs.Db.QueryOne(&user_battery, `UPDATE user_battery SET battery=? WHERE user_id = ?`, steps_battery, user_id)
	if err != nil && err.Error() != "pg: no rows in result set" {
		log.Warn("[AddTime] [%v] Error UPDATE ar_shot_history", user_id, err.Error())

		return 0, err
	}

	log.Infof("[AddTime] [%v] Perc bat %f from seconds %v", user_id, steps_battery, seconds)

	UpdateGameHistory(user_id, 4, &diff, nil)

	return steps_battery, nil
}

func UpdateGameHistory(user_id int, type_ int, battery *float32, block_data *int) {
	var game_history mdels.Game_history
	timein := time.Now().Local().Add(time.Hour * 5)
	game_history.Time = timein
	//location, _ := time.LoadLocation("Europe/Moscow")
	flag, _ := CheckInfBatter(user_id)
	game_history.Inf_battery = flag
	game_history.User_id = user_id
	game_history.Type = type_
	game_history.Battery = battery
	game_history.Block_data = block_data

	_, err := configs.Db.QueryOne(&game_history, `INSERT INTO game_history (user_id, type, battery, block_data, inf_battery, time) VALUES (?user_id, ?type, ?battery, ?block_data, ?inf_battery, ?time)RETURNING id`, game_history)
	if err != nil {
		log.Warn("[UpdateGameHistory] [%v] Error INSERT ", game_history.User_id, err.Error())
	}
}
