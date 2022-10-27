package sqlconf

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
)

type Conf struct {
	db       *sql.DB
	AppName  string
	LogsDir  string
	DBFile   string
	Item     map[string]string
	Logger   *Logger
	Mail     *Mail
	Bar      *progressbar.ProgressBar
	H2Server *H2Server
}

var Config *Conf

var (
	ts_now        int64             = time.Now().Unix()
	defaultConfig map[string]string = make(map[string]string, 10)
	zapLogger     *zap.Logger
)

func NewConf(appName string, appLogsDir string) *Conf {
	Config = &Conf{
		AppName: appName,
		LogsDir: appLogsDir,
	}
	Config.SetLogger()

	defaultConfig["app_first_run"] = strconv.FormatInt(ts_now, 10)
	defaultConfig["app_conf_update"] = strconv.FormatInt(ts_now, 10)
	defaultConfig["app_name"] = "sqlconf"
	defaultConfig["app_author"] = "harryzhu"
	defaultConfig["app_license"] = "MIT"
	defaultConfig["app_version"] = "1.0.0"
	defaultConfig["app_data_dir"] = "./data"
	defaultConfig["app_logs_dir"] = "./logs"
	defaultConfig["app_temp_dir"] = "./temp"

	if Config.DBFile == "" {
		sqlconfenv := strings.ToLower(GetEnv("SQLCONFENV", ""))
		if sqlconfenv != "" {
			Config.DBFile = strings.Join([]string{"./conf", sqlconfenv, "db"}, ".")
		} else {
			Config.DBFile = "./conf.db"
		}
	}

	zapLogger.Info("config file", zap.String("db", Config.DBFile))

	firstRun := false

	_, err := os.Stat(Config.DBFile)
	if err != nil {
		firstRun = true
		zapLogger.Info("First Run")
	}
	Config.Open()

	if firstRun == true {
		Config.LoadData(defaultConfig)
	}
	Config.Refresh().SetLogger()

	Config.RequiredKeys([]string{"app_name", "app_logs_dir"})

	return Config
}

func init() {

}

func (c *Conf) Refresh() *Conf {
	rows, err := c.db.Query("select * from settings")
	if err != nil {
		zapLogger.Fatal("conf-refresh", zap.Error(err))
	}

	var id, name, val string
	var item map[string]string = make(map[string]string)

	for rows.Next() {
		rows.Scan(&id, &name, &val)
		if name != "" {
			item[name] = val
		}
	}
	c.Item = item
	return c
}

func (c *Conf) Open() *Conf {
	db, err := sql.Open("sqlite3", c.DBFile)
	if err != nil {
		zapLogger.Fatal("conf:open", zap.Error(err))
	} else {
		zapLogger.Info("conf:open", zap.String("file", c.DBFile))
	}

	settingsTable := `
	CREATE TABLE IF NOT EXISTS settings(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name CHARACTER(64) UNIQUE NOT NULL,
		val TEXT(65535) NOT NULL DEFAULT "");`

	indexies := `CREATE INDEX IF NOT EXISTS uidxname ON settings ("name");`
	tables := []string{settingsTable, indexies}

	for _, t := range tables {
		_, err = db.Exec(t)
		if err != nil {
			zapLogger.Fatal("db-exec", zap.Error(err))
		}
	}
	c.db = db
	return c
}

func (c *Conf) Close() {
	c.db.Close()
}

func (c *Conf) Set(k, v string) *Conf {
	k = strings.Trim(k, " ")
	v = strings.Trim(v, " ")
	var id, name, val string
	q := "select * from settings where name=?"
	err := c.db.QueryRow(q, k).Scan(&id, &name, &val)

	if err != nil {
		q = "INSERT INTO settings(val,name) VALUES(?,?)"
	} else {
		q = "update settings set val=? where name=?"
	}

	stmt, err := c.db.Prepare(q)
	if err != nil {
		zapLogger.Fatal("set:db-prepare", zap.Error(err))
	}

	_, err = stmt.Exec(v, k)
	if err != nil {
		zapLogger.Fatal("set:stmt-exec", zap.Error(err))
	}

	return c
}

func (c *Conf) Push(k, v string) *Conf {
	k = strings.Trim(k, " ")
	v = strings.Trim(v, " ")
	vlist := StringToSlice(v)

	var id, name, val string
	var valList []string
	q := "select * from settings where name=?"
	err := c.db.QueryRow(q, k).Scan(&id, &name, &val)

	if err != nil {
		q = "INSERT INTO settings(val,name) VALUES(?,?)"
		zapLogger.Info("insert")
	} else {
		q = "update settings set val=? where name=?"
		zapLogger.Info("update")
		valList = StringToSlice(val)

		for _, newVal := range vlist {
			zapLogger.Info("update", zap.String("range valList", newVal))
			valList = append(valList, newVal)
		}

	}

	stmt, err := c.db.Prepare(q)
	if err != nil {
		zapLogger.Fatal("set:db-prepare", zap.Error(err))
	}

	v = SliceToString(SliceUnique(valList))
	zapLogger.Info("update", zap.String("valList", v))

	_, err = stmt.Exec(v, k)
	if err != nil {
		zapLogger.Fatal("set:stmt-exec", zap.Error(err))
	}

	return c
}

func (c *Conf) Delete(k string) *Conf {
	if len(k) == 0 || len(k) > 64 {
		zapLogger.Fatal("--name length must be greater than 0 and less then 64")
	}
	var id, name, val string
	q := "select * from settings where name=?"
	err := c.db.QueryRow(q, k).Scan(&id, &name, &val)

	if err != nil {
		zapLogger.Fatal("delete", zap.Error(err))
	} else {
		q = "delete from settings where name=?"
	}

	stmt, err := c.db.Prepare(q)
	if err != nil {
		zapLogger.Fatal("stmt-prepare", zap.Error(err))
	}
	_, err = stmt.Exec(k)

	if err != nil {
		zapLogger.Fatal("stmt-exec", zap.Error(err))
	} else {
		zapLogger.Info("key was deleted", zap.String("key", k))
	}

	return c
}

func (c *Conf) LoadData(m map[string]string) *Conf {
	for k, v := range m {
		if k != "" {
			c.Set(k, v)
		}
	}

	return c
}

func (c *Conf) SetLogger() *Conf {
	var l *Logger = &Logger{}
	logs_dir := strings.ToLower(c.LogsDir)
	app_name := strings.ToLower(c.AppName)
	if logs_dir == "" {
		logs_dir = "./logs"
	}
	if app_name == "" {
		app_name = "sqlconf"
	}

	c.Logger = l.initLogger(logs_dir, app_name)
	zapLogger = c.Logger.ZapLogger
	return c
}

func (c *Conf) SetBar(max int64, title string) *Conf {
	c.Bar = initProgressBar(max, title)
	return c
}

func (c *Conf) SetH2Server() *Conf {
	c.H2Server = h2server
	return c
}

func (c *Conf) SetMail() *Conf {
	var m *Mail = &Mail{}
	m.WithSMTPEnv("SQLCONFSMTPHOST", "SQLCONFSMTPPORT", "SQLCONFSMTPUSERNAME", "SQLCONFSMTPPASSWORD")
	m.WithMailEnv("SQLCONFSMTPFROM", "SQLCONFSMTPTO", "SQLCONFSMTPCC", "SQLCONFSMTPBCC")

	c.Mail = m
	return c
}

func (c *Conf) Print() {
	log.Println("===== Print All Config Items =====\n")
	var cItemKeys []string
	for k, _ := range c.Item {
		cItemKeys = append(cItemKeys, k)
	}

	sort.Strings(cItemKeys)

	for _, k := range cItemKeys {
		fmt.Println(k, "=", c.Item[k])
	}
	fmt.Println("")
	log.Println("===== Config END =====")
}

func (c *Conf) ToString(s string) string {
	if _, ok := c.Item[s]; ok {
		return c.Item[s]
	} else {
		log.Println("***ERROR***::item does not exist in conf: ", s)
	}

	return ""
}

func (c *Conf) ToBool(s string) bool {
	itm := strings.ToLower(c.Item[s])
	if itm == "false" || itm == "0" {
		return false
	}
	return true
}

func (c *Conf) ToInt(s string) int {
	res, err := strconv.Atoi(c.Item[s])
	if err != nil {
		zapLogger.Fatal("ToInt", zap.Error(err))
	}
	return res
}

func (c *Conf) ToInt64(s string) int64 {
	res, err := strconv.ParseInt(c.Item[s], 10, 64)
	if err != nil {
		zapLogger.Fatal("ToInt64", zap.Error(err))
	}
	return res
}

func (c *Conf) ToFloat64(s string) float64 {
	res, err := strconv.ParseFloat(c.Item[s], 64)
	if err != nil {
		zapLogger.Fatal("ToFloat64", zap.Error(err))
	}
	return res
}

func (c *Conf) ToError(s string) (err error) {
	return errors.New(c.Item[s])
}

func (c *Conf) RequiredKeys(m []string) {
	for _, k := range m {
		if c.Item[k] == "" {
			zapLogger.Fatal("cannot be empty, pls use ./sqlconfctl set --name= --val= ", zap.String("name", k))
		}
	}
}

func (c *Conf) AlwaysPostRun() error {
	c.Logger.MailReportCurrentError()

	return nil
}
