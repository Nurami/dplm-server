package main

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

func connectToDB() {
	connStr := "user=postgres password=postgres dbname=carwashing sslmode=disable host=localhost port=5432"
	dataBase, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Panic(err)
	}
	db = dataBase
}

func getAgentsIDs() []string {
	rows, err := db.Query("SELECT DISTINCT agent_id FROM log")
	if err != nil {
		log.Error(err)
	}
	result := make([]string, 0)
	for rows.Next() {
		tmp := ""
		err = rows.Scan(&tmp)
		if err != nil {
			log.Error(err)
		}
		result = append(result, tmp)
	}
	return result
}

func getDevicesNames() []string {
	rows, err := db.Query("SELECT DISTINCT device.name FROM log INNER JOIN device ON log.device_id=device.id")
	if err != nil {
		log.Error(err)
	}
	result := make([]string, 0)
	for rows.Next() {
		tmp := ""
		err = rows.Scan(&tmp)
		if err != nil {
			log.Error(err)
		}
		result = append(result, tmp)
	}
	return result
}

func getLogsLevels() []string {
	rows, err := db.Query("SELECT DISTINCT level FROM log")
	if err != nil {
		log.Error(err)
	}
	result := make([]string, 0)
	for rows.Next() {
		tmp := ""
		err = rows.Scan(&tmp)
		if err != nil {
			log.Error(err)
		}
		result = append(result, tmp)
	}
	return result
}

func writeToDBlogInfo(logInfo logInfo) {
	_, err := db.Exec("INSERT INTO log(agent_id, time, function, level, id, message, device_id) VALUES ($1, $2, $3, $4, $5, $6, $7)", logInfo.agentID, logInfo.time, logInfo.function, logInfo.level, logInfo.id, logInfo.message, logInfo.deviceID)
	if err != nil {
		log.Error(err)
	}
}

func getResultLogInfos(clRe clientRequest) []byte {
	tmp1 := strings.Replace(clRe.FirstDate, "T", " ", -1)
	tmp2 := strings.Replace(clRe.SecondDate, "T", " ", -1)
	firstTime, err := time.Parse("2006-01-02 15:04", tmp1)
	if err != nil {
		log.Error(err)
	}
	secondTime, err := time.Parse("2006-01-02 15:04", tmp2)
	if err != nil {
		log.Error(err)
	}
	rows, err := db.Query("SELECT log.agent_id, log.level, log.time, log.message, device.name FROM log INNER JOIN device ON device.name=$1 WHERE (log.time BETWEEN $2 AND $3) AND log.level=$4 AND log.agent_id=$5", clRe.DeviceType, firstTime, secondTime, clRe.Level, clRe.AgentID)
	if err != nil {
		log.Error(err)
	}
	resultInfos := make([]resultInfo, 0)
	for rows.Next() {
		tmp := resultInfo{}
		err = rows.Scan(&tmp.AgentID, &tmp.Level, &tmp.Time, &tmp.Message, &tmp.Name)
		if err != nil {
			log.Error(err)
		}
		resultInfos = append(resultInfos, tmp)
	}
	result, err := json.Marshal(resultInfos)
	if err != nil {
		log.Error(err)
	}
	return result
}
