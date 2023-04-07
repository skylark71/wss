package logic

import (
	"fmt"
	"math/rand"
	"single-win-system/configs"
	mdels "single-win-system/models"
	"single-win-system/util"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var probabilities map[int]float32
var coefficients map[int]float32

const (
	cells_count   = 249
	SLOT_MACHINE  = 1
	AR_MACHINE    = 2
	BONUS_MACHINE = 3
)

func GetProbabilities(type_ int, user_id int) map[int]float32 {

	var str string
	if type_ == SLOT_MACHINE {
		str = "sm_cfg"
	} else if type_ == AR_MACHINE {
		str = "ar_cfg"
	} else {
		str = "bd_cfg"
	}

	var ar_cfgs []mdels.Ar_cfg
	_, err := configs.Db.QueryOne(&ar_cfgs, `SELECT * FROM `+str+` ORDER BY id ASC`)
	if err != nil && err.Error() != "pg: multiple rows in result set" {
		log.Warn("[GetProbabilities] Error SELECT ", err.Error())
	}
	var tmp []float32
	for _, v := range ar_cfgs {
		tmp = append(tmp, v.Probabilities)
	}

	log.Info("[GetProbabilities]", tmp)

	if type_ == AR_MACHINE {
		var user_monthly_win_history mdels.User_monthly_win_history
		user_monthly_win_history = GetMonthlyWinHistory(user_id)

		if user_monthly_win_history.Amount >= user_monthly_win_history.Win_limit {
			for i := 0; i < len(tmp); i++ {
				tmp[i] = 0
			}
		} else if user_monthly_win_history.Amount >= 25000 {
			for i := 0; i < len(tmp); i++ {
				if i < 5 {
					tmp[i] = 0
				}
			}
		}
	}

	if type_ == BONUS_MACHINE {
		var user_monthly_win_history mdels.User_monthly_win_history
		user_monthly_win_history = GetMonthlyWinHistory(user_id)

		if user_monthly_win_history.Amount >= user_monthly_win_history.Win_limit {
			for i := 0; i < len(tmp); i++ {
				if i < 9 {
					tmp[i] = 0
				}
			}
		} else if user_monthly_win_history.Amount >= 25000 {
			for i := 0; i < len(tmp); i++ {
				if i < 7 {
					tmp[i] = 0
				}
			}
		}
	}

	log.Info("[GetProbabilities]", tmp)

	keyval := make(map[int]float32)
	for i := 0; i < len(tmp); i++ {
		keyval[i+1] = tmp[i]
	}
	return keyval
}

func GetCoefficients(type_ int) map[int]float32 {

	var str string
	if type_ == SLOT_MACHINE {
		str = "sm_cfg"
	} else if type_ == AR_MACHINE {
		str = "ar_cfg"
	} else {
		str = "bd_cfg"
	}

	var ar_cfgs []mdels.Ar_cfg
	_, err := configs.Db.QueryOne(&ar_cfgs, `SELECT * FROM `+str)
	if err != nil && err.Error() != "pg: multiple rows in result set" {
		log.Warn("[GetCoefficients] Error SELECT ", err.Error())
	}
	var tmp []float32
	for _, v := range ar_cfgs {
		tmp = append(tmp, v.Coefficients)
	}

	log.Info("[GetCoefficients]", tmp)

	keyval := make(map[int]float32)
	for i := 0; i < len(tmp); i++ {
		keyval[i+1] = tmp[i]
	}
	return keyval
}

func geneateReeplsMap(type_ int, user_id int) []int {

	probabilities = GetProbabilities(type_, user_id)

	last := probabilities[len(probabilities)-1] // C
	if last == 0 {
		a := []int{0, 0, 0}
		return a
	}

	var reelmap map[int]interface{}
	var a []int
	for key, element := range probabilities {
		reelmap := util.ArrayFill(len(reelmap), int(element), key)
		for _, v := range reelmap {
			vv := v.(int)
			a = append(a, int(vv))
		}
	}

	log.Info("[geneateReeplsMap] probabilities ", probabilities)

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })

	first := a[0]
	a = a[1:]
	a = append(a, first)

	return a
}

func generateCombination(a []int) []int {
	var combination []int
	var i int
	for i = 0; i < 3; i++ {
		combination = append(combination, a[rand.Intn(len(a))])
	}

	winProbabilities := []int{4, 5, 6, 7, 8}
	if util.Contains(winProbabilities, combination[0]) {
		winProbabilities = []int{1, 2, 3, 5, 7, 9}
	} else {
		winProbabilities = []int{1, 3, 5, 7}
	}

	min := 0
	max := 9
	rand.Seed(time.Now().UnixNano())
	rand := rand.Intn(max-min) + min
	var comb []int
	if util.Contains(winProbabilities, rand) {
		for i = 0; i < 3; i++ {
			comb = append(comb, combination[0])
		}
		return comb
	}

	return combination
}

func isWin(comb []int) bool {
	first := comb[0]
	for _, v := range comb {
		if v != first {
			return false
		}
	}
	return true
}

func processWin(comb []int, bet int, type_ int) float32 {
	coefficients := GetCoefficients(type_)
	combinationType := comb[0]
	var coef float32
	for key, element := range coefficients {
		if key == combinationType {
			coef = element
		}
	}
	return float32(bet) * coef
}

func processWinBonus(comb []int, index int, bet int, type_ int) float32 {
	coefficients := GetCoefficients(type_)
	combinationType := comb[index]
	var coef float32
	for key, element := range coefficients {
		if key == combinationType {
			coef = element
		}
	}
	return float32(bet) * coef
}

func Spin(bet int, type_ int, user_id int) (bool, float32, []int) {
	prize := float32(0)
	reelsMap := geneateReeplsMap(type_, user_id)
	combination := generateCombination(reelsMap)
	isWin := isWin(combination)

	if isWin {
		prize = processWin(combination, bet, type_)
	}

	return isWin, prize, combination
}

func SpinBonus(bet int, type_ int, user_id int) float32 {
	prize := float32(0)
	reelsMap := geneateReeplsMap(type_, user_id)
	randomIndex := rand.Intn(len(reelsMap))
	prize = processWinBonus(reelsMap, randomIndex, bet, type_)
	return prize
}

func getBonus(array []int) int {
	randomIndex := rand.Intn(len(array))
	pick := array[randomIndex]
	return pick
}

func GetEkbTime() time.Time {
	t := time.Now()
	loc := time.FixedZone("UTC+5", 5*60*60)
	//location, _ := time.LoadLocation("Europe/Moscow")
	return t.In(loc)
}

func getFirstDayOfMonth() time.Time {
	date := time.Now()
	return date.AddDate(0, 0, -date.Day()+1)
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}

func ArrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(fmt.Sprint(a), " ", delim, -1), "[]")
}
