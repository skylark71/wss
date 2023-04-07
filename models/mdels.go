package mdels

import (
	"time"
)

type Obscure struct {
	Hp      int
	Armour  int
	Battery int
}

type Data struct {
	Battery   int
	Blockdata int
	Prize     float32
}

type bd_cfg struct {
	id            int
	probabilities float32
	coefficients  float32
}

type User struct {
	Id                   int
	Number               string
	Balance              int
	Lang                 string
	Fcm_token            string
	Password_hash        string
	Password_reset_token string
	Email                string
	Status               int
	Created_at           int
	Updated_at           int
	Shots                int
	Name                 string
	Verification_code    string
	Is_tutorial_finished int
	Photo                string
	Mobile_operator_id   int
	Instagram_id         string
	Facebook_id          string
	Timezone             int
	Tutorial_finished_at time.Time
	Token_balance        int
	Steps_ban            bool
	Last_online          time.Time
	Score                int
	Tutorial_mt          bool
}

type User_tokens struct {
	Id              int
	User_id         int
	Access_token    string
	Network         string
	Network_user_id string
	Created_at      string
}

type Ar_shoot_history struct {
	Id         int
	User_id    int
	Place_id   string
	Try        int
	Prize      int
	Created_at time.Time
	Updated_at time.Time
	Expired_at time.Time
	Place_type int
}

type Balance_history struct {
	Id                int
	Amount            float32
	User_id           int
	Date              time.Time
	Type_id           int
	Balance           float32
	Linked_history_id *int
}

type User_monthly_win_history struct {
	Id        int
	User_id   int
	Amount    int
	Date      time.Time
	Win_limit int
}

type Ar_cfg struct {
	Id            int
	Probabilities float32
	Coefficients  float32
}

type Slot_configuration struct {
	Id                 int
	Min_bet            int
	Max_bet            int
	Bet_step           int
	Token_combinations string
}

type Slot_history struct {
	Id          int
	User_id     int
	Bet         int
	Prize       *int
	Combination string
	Created_at  time.Time
}

type Bonus_history struct {
	Id       int
	User_id  int
	Date     time.Time
	Amount   int
	Datetime time.Time
}

type Bonus_events struct {
	Id                int
	Notification_text string
	Type_id           int
	Header            string
}

type Game_tokens_configuration struct {
	Id                          int
	Ar_incrementation_percent   int
	Slot_incrementation_percent int
	Created_at                  time.Time
	Updated_at                  time.Time
}

type Cfg_game struct {
	Id                   int
	Battery              int
	Storage              int
	Battery_armour       string
	Battery_life         string
	Obscure_life         string
	Obscure_storage      string
	Obscure_charge       string
	Block_data_sm        string
	Battery_time_charge  float32
	Battery_steps_charge float32
	Is_active            bool
	Bonus_battery        string
	Bonus_battery_per    int
	Tutorial_lvls        string
	Cool_down            string
	Switcher             bool
}

type Settings_obscure struct {
	Up   int `json:"up,string,omitempty"`
	Down int `json:"down,string,omitempty"`
}

type Fight struct {
	Obscure         Obscure
	Battery_armour  int
	Battery_life    int
	Obscure_charge  int
	Obscure_storage int
}

type User_battery struct {
	Id          int
	User_id     int
	Battery     float32
	Block_data  float32
	Bonus_start time.Time
	Bonus       bool
	Bonus_end   time.Time
}

type Spin_response struct {
	Iswin       bool
	Prize       float32
	Combination []int
}

type Bonus_response struct {
	Prize float32
}

type Game_history struct {
	Id          int
	User_id     int
	Type        int
	Battery     *float32
	Block_data  *int
	Inf_battery bool
	Time        time.Time
}
