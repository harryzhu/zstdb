package sqlconf

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/zap"
)

type Conf struct {
	db     *sql.DB
	Item   map[string]string
	Logger *zap.Logger
	Bar    *progressbar.ProgressBar
}

func (c *Conf) Refresh() *Conf {
	rows, err := c.db.Query("select * from settings")
	if err != nil {
		log.Fatal(err)
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

func (c *Conf) Open(f string) *Conf {
	db, err := sql.Open("sqlite3", f)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
	}
	c.db = db
	return c
}

func (c *Conf) Close() {
	c.db.Close()
}

func (c *Conf) Set(k, v string) *Conf {
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
		log.Fatal(err)
	}

	_, err = stmt.Exec(v, k)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func (c *Conf) Delete(k string) *Conf {
	if len(k) == 0 || len(k) > 64 {
		log.Fatal("--name length must be greater than 0 and less then 64")
	}
	var id, name, val string
	q := "select * from settings where name=?"
	err := c.db.QueryRow(q, k).Scan(&id, &name, &val)

	if err != nil {
		log.Fatal(err)
	} else {
		q = "delete from settings where name=?"
	}

	stmt, err := c.db.Prepare(q)
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec(k)

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("key(" + k + ") was deleted")
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

func (c *Conf) SetLogger(log_dir, app_name string) *Conf {
	initLogger(log_dir, app_name)
	c.Logger = logger

	return c
}

func (c *Conf) SetBar(max int64, title string) *Conf {
	c.Bar = initProgressBar(max, title)
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
	log.Println("===== END =====")
}

func (c *Conf) ToString(s string) string {
	if _, ok := c.Item[s]; ok {
		return c.Item[s]
	}

	return ""
}

func (c *Conf) ToInt(s string) int {
	res, err := strconv.Atoi(c.Item[s])
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func (c *Conf) ToInt64(s string) int64 {
	res, err := strconv.ParseInt(c.Item[s], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func (c *Conf) ToFloat64(s string) float64 {
	res, err := strconv.ParseFloat(c.Item[s], 64)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func (c *Conf) ToError(s string) (err error) {
	return errors.New(c.Item[s])
}

func (c *Conf) FatalEmpty(m []string) {
	for _, k := range m {
		if c.Item[k] == "" {
			log.Fatalf("%v cannot be empty, pls use ./confctl set --name=%v --val=...", k, k)
		}
	}
}
