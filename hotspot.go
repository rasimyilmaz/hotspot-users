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

func serve(closesignal chan int) {
	flag.Parse()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/", QueryQuests)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	<-closesignal
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
	}
	select {
	case <-ctx.Done():
		srv.Close()
	}
}
func serveHTTP() {
	http.Handle("/", http.FileServer(http.Dir(".")))
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
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s", *server, *user, *password, *port, name)
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

	stmt, err := conn.Prepare("SELECT Q.CITARIHI,Q.COUTTARIHI,CASE WHEN LEN(M.VATNO)=11 then M.VATNO WHEN M.KIMLIKTURU='PASAPORT' OR M.KIMLIKTURU='DİĞER' THEN ISNULL(M.KIMLIKNO,'YABANCI') ELSE 'NOTSENSE' END as KIMLIKNO,ISNULL(M.KIMLIKTURU,'') AS KIMLIKTURU  ,ISNULL(M.AD,'') +' ' + SUBSTRING(ISNULL(M.SOYAD,''),0,2)+'.' AS ADI,ISNULL(YEAR(M.DOGUMTARIHI),0) AS DOGUMYILI FROM QRES_ALL AS Q INNER JOIN REZKISI R ON R.KNO = Q.KNO LEFT JOIN MISAFIR M ON M.MID = R.MID where Q.DURUM='I' AND NOT Q.ODANO LIKE 'T%' and ( M.KIMLIKTURU is not null or M.VATNO is not null) and M.DOGUMTARIHI is not null")
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
