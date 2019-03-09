package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
)

var (
	debug    = flag.Bool("debug", false, "enable debugging")
	password = flag.String("password", "admin1982", "the database password")
	port     = flag.Int("port", 1433, "the database port")
	server   = flag.String("server", "SRV", "the database server")
	user     = flag.String("user", "adminakad", "the database user")
)

//Guest type contains information check in date , check out date , type of id card , id number, name , birth year
type Guest struct {
	CheckInDate  time.Time `json:"check_in_date"`
	CheckOutDate time.Time `json:"check_out_date"`
	IDCard       string    `json:"id_card"`
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	BirthYear    int       `json:"birth_year"`
}

func main() {
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(".")))

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/", QueryQuests)
	r.Run()
	http.ListenAndServe(":3000", nil)
}

//QueryQuests Responses Quests with JSON format
func QueryQuests(c *gin.Context) {
	response, err := GetGuests(c.Query("name"))
	if err != nil {
		c.JSON(500, gin.H{
			"message": err.Error(),
		})
	} else {
		c.JSON(200, response)
	}
}

//GetGuests Input parameter name is stand for database name , return value slice of guest stuct and if there is no error , error is nil
func GetGuests(name string) ([]Guest, error) {
	ctx := context.Background()
	var Quests []Guest
	if *debug {
		fmt.Printf(" password:%s\n", *password)
		fmt.Printf(" port:%d\n", *port)
		fmt.Printf(" server:%s\n", *server)
		fmt.Printf(" user:%s\n", *user)
	}
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s", *server, *user, *password, *port, name)
	if *debug {
		fmt.Printf(" connString:%s\n", connString)
	}
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Print("Open connection failed:", err.Error())
		return Quests, err
	}
	err = conn.PingContext(ctx)
	if err != nil {
		log.Print(err.Error())
		return nil, err
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT Q.CITARIHI,Q.COUTTARIHI,CASE WHEN M.KIMLIKTURU='PASAPORT' THEN ISNULL(M.KIMLIKNO,'12345678901') ELSE ISNULL(M.VATNO,'12345678901') END as KIMLIKNO,ISNULL(M.KIMLIKTURU,'') AS KIMLIKTURU  ,M.AD +' ' + SUBSTRING(M.SOYAD,0,2)+'.' AS ADI,YEAR(M.DOGUMTARIHI) AS DOGUMYILI FROM QRES_ALL AS Q INNER JOIN REZKISI R ON R.KNO = Q.KNO LEFT JOIN MISAFIR M ON M.MID = R.MID where Q.DURUM='I' AND NOT Q.ODANO LIKE 'T%'")
	if err != nil {
		log.Print("Prepare failed:", err.Error())
		return Quests, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		log.Print(err)
		return Quests, err
	}
	defer rows.Close()
	for rows.Next() {
		var row Guest
		err := rows.Scan(&row.CheckInDate, &row.CheckOutDate, &row.ID, &row.IDCard, &row.Name, &row.BirthYear)
		if err != nil {
			log.Print(err)
		}
		Quests = append(Quests, row)
	}
	err = rows.Err()
	if err != nil {
		log.Print(err)
		return Quests, err
	}
	return Quests, nil
}
