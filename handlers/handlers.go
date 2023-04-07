package handlers

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"single-win-system/configs"
	"single-win-system/logic"
	mdels "single-win-system/models"
	"strings"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Test struct {
	Hp     int
	Armour int
}

type User struct {
	Battery   int
	Blockdata int
}

type Win struct {
	Battery_win   int
	Blockdata_win int
	Prize         float32
}

type Arshoot struct {
	Place_id       string
	Place_location int
	User_location  int
	Place_type     int
	Game_token     int
}

func Router() *mux.Router {
	isReady := &atomic.Value{}
	isReady.Store(true)
	r := mux.NewRouter()
	r.HandleFunc("/healthz", healthz)
	r.HandleFunc("/", sendEpoch)
	r.HandleFunc("/readyz", readyz(isReady))
	r.HandleFunc("/ws", getResultHandler)

	return r
}

func getResultHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	h := http.Header{}
	h.Set("Sec-Websocket-Protocol", websocket.Subprotocols(r)[0])

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	var user mdels.User
	log.Infof("access_token %s", websocket.Subprotocols(r)[0])
	_, err = configs.Db.QueryOne(&user, `SELECT "user".* FROM "user" JOIN user_tokens ON user_tokens.user_id = "user".id WHERE user_tokens.access_token = ?`, websocket.Subprotocols(r)[0])
	if err != nil {
		log.Warn("DB Error %s", err.Error())
		//ws.Close()
	}
	log.Infof("User %v", user)
	if user.Status == logic.USER_BLOCKED || user.Status == logic.USER_DELETED {
		//return
	}
	upgrader.Subprotocols = append(upgrader.Subprotocols, r.Header.Get("Sec-Websocket-Protocol"))

	log.Infof("Client connected: %s", ws.RemoteAddr().String())

	var fight_global mdels.Fight
	var empty_data mdels.Data
	count := 0
	var arShoot Arshoot
	for {
		// ws.SetReadDeadline(time.Now().Add(45 * time.Second))
		// ws.SetWriteDeadline(time.Now().Add(45 * time.Second))
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			// if err.Error() == "websocket: close 1005 (no status)" || err.Error() == "websocket: close 1006 (abnormal closure): unexpected EOF" {
			// 	continue
			// }
			if ce, ok := err.(*websocket.CloseError); ok {
				switch ce.Code {
				case websocket.CloseGoingAway,
					websocket.CloseNoStatusReceived,
					websocket.CloseAbnormalClosure:
					ws.Close()
					log.Warn("[Web socket code] code %v", ce.Code)
					return
				case websocket.CloseNormalClosure:
					ws.Close()
					return
				}
			}

			configs.Db.Close()
			log.Warn("[Ошибка получения текста из ws] message%s", err.Error())
			return
		}

		// ws.SetReadDeadline(time.Now().Add(15 * time.Minute))
		// ws.SetWriteDeadline(time.Now().Add(15 * time.Minute))

		// if err != nil || messageType == websocket.CloseMessage {
		// 	break
		// }

		//!Обновление заярда батареи в сек
		if string(message) == "sec" {
			battery, err := logic.AddTime(60, user.Id)
			if err != nil {
				log.Info(err)
			}

			if battery >= 100.0 {
				battery = 100
			}

			s := fmt.Sprintf("%f", battery)

			if err := ws.WriteMessage(messageType, []byte(s)); err != nil {
				log.Info(err)
				return
			}
		}

		//!Получаем обскура
		if strings.Contains(string(message), "place_id") {
			json.Unmarshal(message, &arShoot)
			fight_global = logic.GetObscure(user.Id)

			json_data, err := json.Marshal(Test{fight_global.Obscure.Hp, fight_global.Obscure.Armour})
			if err != nil {
				log.Warn("[Handler place_id] Error: ", err.Error())
				if err := ws.WriteMessage(messageType, []byte(err.Error())); err != nil {
					log.Info(err)
					return
				}
			}
			if err := ws.WriteMessage(messageType, json_data); err != nil {
				log.Warn("[Handler place_id] Error: ", err.Error())
				return
			}
		}

		//!Получаем батарйеку
		if string(message) == "battery" {
			data, err := logic.GetUpdateBattery(user.Id)
			if err != nil {
				log.Warn("[Handler battery] Error: ", err.Error())
			}
			json_data, err := json.Marshal(math.Round(float64(data)))
			if err != nil {
				log.Warn("[Handler battery] Error: ", err.Error())
			}

			if err := ws.WriteMessage(messageType, json_data); err != nil {
				log.Warn("[Handler battery] Error: ", err.Error())
				return
			}
		}

		//!Остаток минут до конца бесконечной батарейки
		if string(message) == "inf_battery" {
			data, err := logic.GetInfBattery(user.Id)
			if err != nil {
				log.Warn("[Handler inf_battery] Error: ", err.Error())
			}
			json_data, err := json.Marshal(data)
			if err != nil {
				log.Warn("[Handler inf_battery] Error: ", err.Error())
			}

			if err := ws.WriteMessage(messageType, json_data); err != nil {
				log.Warn("[Handler inf_battery] Error: ", err.Error())
				return
			}
		}

		//!Перестрелка с обскуром
		if strings.Contains(string(message), "Hp") {
			logic.GetUpdateBattery(user.Id)

			fight, data, err := logic.ArShootReq(count, fight_global, user, arShoot.Place_id, arShoot.Place_location, arShoot.User_location, arShoot.Place_type, arShoot.Game_token)
			if err != nil {
				//logic.UpdateBattery(0, user.Id)
				if err := ws.WriteMessage(messageType, []byte(err.Error())); err != nil {
					count = 0
					log.Info(err)
					return
				}
			}

			json_data, err := json.Marshal(mdels.Obscure{fight.Obscure.Hp, fight.Obscure.Armour, fight.Obscure.Battery})
			if err != nil {
				log.Fatal(err)
			}
			if data != empty_data {
				count = 0
				if err := ws.WriteMessage(messageType, json_data); err != nil {
					log.Fatal(err)
					return
				}

				if arShoot.Place_id == "0" {
					logic.SetTutorialAr(user)
				}

				if err := ws.WriteMessage(messageType, []byte("Выиграл")); err != nil {
					log.Info(err)
					return
				}

				json_win, err := json.Marshal(Win{data.Battery, data.Blockdata, data.Prize})
				if err = ws.WriteMessage(messageType, json_win); err != nil {
					log.Warn(err)
					return
				}

			} else {
				if err := ws.WriteMessage(messageType, json_data); err != nil {
					log.Fatal(err)
					return
				}
				count++
			}

			fight_global = fight

		}
	}

}

func float64ToByte(f float32) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, f)
	if err != nil {
		log.Warn("[float64ToByte] binary.Write failed:", err.Error())
	}
	return buf.Bytes()
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readyz(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func sendEpoch(w http.ResponseWriter, r *http.Request) {
	// Для отладки:
	tmpl, _ := template.ParseFiles("templates/index.html")
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	//w.WriteHeader(http.StatusOK)
}
