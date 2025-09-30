package models

type LoginLog struct {
	UUIDModel

	ID       uint   `gorm:"autoIncrement;type:integer;primaryKey" json:"id"`
	UserID   int    `json:"user_id" gorm:"type:int;index"`
	Username string `json:"username" gorm:"type:varchar(100);default:'';index"`
	IP       string `json:"ip" gorm:"type:varchar(64);default:'';"`
	UA       string `json:"ua" gorm:"type:varchar(512);default:'';"`
}

func AddLoginLog(userID int, username string, ip string, ua string) error {
	log := LoginLog{UserID: userID, Username: username, IP: ip, UA: ua}
	return db.Create(&log).Error
}

func GetRecentLoginLogs(limit int) ([]LoginLog, error) {
	if limit <= 0 {
		limit = 8
	}
	var logs []LoginLog
	err := db.Model(&LoginLog{}).Order("id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}
