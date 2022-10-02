package sqlconf

import (
	"log"

	"database/sql"
	"sort"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Conf struct {
	db   *sql.DB
	Item map[string]string
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
	name VARCHAR(64) UNIQUE NOT NULL,
	val VARCHAR(1024) NOT NULL DEFAULT "");`

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
		return c
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

func (c *Conf) Print() {
	log.Println("===== Print All Items =====")
	var cItemKeys []string
	for k, _ := range c.Item {
		cItemKeys = append(cItemKeys, k)
	}

	sort.Strings(cItemKeys)

	for _, k := range cItemKeys {
		log.Println(k, "=", c.Item[k])
	}
	log.Println("===== END =====")
}

func ToInt(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func ToInt64(s string) int64 {
	res, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	return res
}

func ToFloat64(s string) float64 {
	res, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatal(err)
	}
	return res
}
