package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/gin-gonic/gin"
)

var (
	debug         = flag.Bool("debug", false, "enable debugging")
	password      = flag.String("password", "admin1982", "the database password")
	port     *int = flag.Int("port", 1433, "the database port")
	server        = flag.String("server", "SRV", "the database server")
	user          = flag.String("user", "adminakad", "the database user")
)

type guest struct {
	Check_in_date  time.Time `json:"check_in_date"`
	Check_out_date time.Time `json:"check_out_date"`
	Id_card        string    `json:"id_card"`
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	Birth_year     int       `json:"birth_year"`
}

func main() {
	flag.Parse()
	r := gin.Default()
	r.GET("/", QueryQuests)
	r.Run()
}
func QueryQuests(c *gin.Context) {
	c.JSON(200, GetGuests(c.Query("name")))
}
func GetGuests(name string) []guest {
	var Quests []guest
	if *debug {
		fmt.Printf(" password:%s\n", *password)
		fmt.Printf(" port:%d\n", *port)
		fmt.Printf(" server:%s\n", *server)
		fmt.Printf(" user:%s\n", *user)
	}
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s", *server, *user, *password, *port, name)
	fmt.Printf(" connString:%s\n", connString)
	conn, err := sql.Open("mssql", connString)
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	defer conn.Close()

	stmt, err := conn.Prepare("SELECT Q.CITARIHI,Q.COUTTARIHI,CASE WHEN M.KIMLIKTURU='PASAPORT' THEN M.KIMLIKNO ELSE M.VATNO END as KIMLIKNO,ISNULL(M.KIMLIKTURU,'') AS KIMLIKTURU  ,M.AD +' ' + SUBSTRING(M.SOYAD,0,2)+'.' AS ADI,YEAR(M.DOGUMTARIHI) AS DOGUMYILI FROM QRES_ALL AS Q INNER JOIN REZKISI R ON R.KNO = Q.KNO LEFT JOIN MISAFIR M ON M.MID = R.MID where Q.DURUM='I' AND NOT Q.ODANO LIKE 'T%'")
	if err != nil {
		log.Fatal("Prepare failed:", err.Error())
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var row guest
		err := rows.Scan(&row.Check_in_date, &row.Check_out_date, &row.Id, &row.Id_card, &row.Name, &row.Birth_year)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(row)
		Quests = append(Quests, row)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return Quests
}
