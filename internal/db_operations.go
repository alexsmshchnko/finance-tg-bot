package internal

import (
	"context"
	"log"
	"time"

	pgx "github.com/jackc/pgx/v5"
)

// urlExample := "postgres://username:password@localhost:5432/database_name"
const DB_URL = "postgres://postgres:postgres@localhost:5433/base"

type user struct {
	username string
}

func NewUser(username string) *user {
	return &user{username: username}
}

type finRecord struct {
	time        time.Time
	category    string
	amount      int
	description string
	msgId       string
	clientId    user
}

func NewFinRec(category string, amnt int, descr string, msgId string) *finRecord {
	return &finRecord{
		time:        time.Now(),
		category:    category,
		amount:      amnt,
		description: descr,
		msgId:       msgId,
		//clientId:    clientId,
	}
}

var dbChan chan interface{}

func init() {
	dbChan = make(chan interface{})

	go dbOp(dbChan)
}

func newConnection() (conn *pgx.Conn, err error) {
	conn, err = pgx.Connect(context.Background(), DB_URL)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
	// 	os.Exit(1)
	// }
	return
}

func dbOp(c chan interface{}) {
	conn, err := newConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	var result interface{}

	for s := range c {
		switch v := s.(type) {
		case user:
			err = conn.QueryRow(context.Background(), "select external_system_token from base.public.user where username = $1", v.username).Scan(&result)
			if err != nil {
				c <- err
			}
			c <- result
		case finRecord:
			func(sqlQuery string) {
				tx, err := conn.Begin(context.Background())
				if err != nil {
					c <- err
				}
				defer tx.Rollback(context.Background())

				_, err = tx.Exec(context.Background(), sqlQuery, v.time, v.category, v.amount, v.description, v.msgId, v.clientId.username)
				if err != nil {
					c <- err
					return
				}

				c <- tx.Commit(context.Background())
			}("INSERT INTO base.public.doc (posting_date, cat, posting_amount, add_info, external_id, client_id) VALUES($1, $2, $3, $4, $5, $6);")
		}

	}
}

func (u *user) GetUserToken() (token string, err error) {
	dbChan <- *u
	data := <-dbChan

	switch v := data.(type) {
	case string:
		token = v
	case error:
		err = v
	default:
	}

	return
}

func (u *user) NewExpense(rec *finRecord) (err error) {
	rec.clientId = *u
	dbChan <- *rec
	data := <-dbChan

	switch v := data.(type) {
	case error:
		err = v
	default:
	}
	return
}

func WriteNewExpense(username string, rec *finRecord) (err error) {
	conn, _ := newConnection()
	defer conn.Close(context.Background())

	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "INSERT INTO base.public.doc (posting_date, cat, posting_amount, add_info, external_id, client_id) VALUES($1, $2, $3, $4, $5, $6);",
		rec.time, rec.category, rec.amount, rec.description, rec.msgId, rec.clientId.username)
	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	return

}

//func connect() {

// 	//conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
// 		os.Exit(1)
// 	}
// 	defer conn.Close(context.Background())

// 	var username string
// 	var external_system_name string
// 	err = conn.QueryRow(context.Background(), "select u.username, u.external_system_name from base.public.user u where u.username = $1", "quile17").Scan(&username, &external_system_name)
// 	if err != nil {
// 		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
// 		os.Exit(1)
// 	}

// 	fmt.Println(username, external_system_name)
// }
