package main

import (
    "crypto/rand"
	"fmt"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/jinzhu/gorm"
    _ "github.com/mattn/go-sqlite3"

)

type Users struct {
    gorm.Model
    Account string `gorm:"primary_key" form:"account" json:"account"`
    Passwd  string `gorm:"not null" form:"passwd" json:"passwd"`
    Name string `gorm:"not null" form:"name" json:"name"`
    Books []Books `gorm:"AssociationForeignKey:UserID" form:"books" json:"books"`
}
type Books struct {
    gorm.Model
    UserID int `form:"userID" json:"userID"`
    Name string `gorm:"not null" form:"name" json:"name"`
    Pages  int `gorm:"not null" form:"pages" json:"pages"`
    Records []BookRecords `gorm:"AssociationForeignKey:BookID" form:"records" json:"records"`
}

type BookRecords struct {
    gorm.Model
    BookID int
    pages int `gorm:"not null" form:"pages" json:"pages"`
}

func InitDb() *gorm.DB {
    db, err := gorm.Open("sqlite3", "./data.db")
    db.LogMode(true)

    if err != nil {
        panic(err)
    }
    if !db.HasTable(&Users{}) {
        db.CreateTable(&Users{})
        db.Set("gorm:table_options","ENGINE=InnoDB").CreateTable(&Users{})
    }
    if !db.HasTable(&Books{}) {
        db.CreateTable(&Books{})
        db.Set("gorm:table_options","ENGINE=InnoDB").CreateTable(&Books{})
    }
    if !db.HasTable(&BookRecords{}) {
        db.CreateTable(&BookRecords{})
        db.Set("gorm:table_options","ENGINE=InnoDB").CreateTable(&BookRecords{})
    }
    return db
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}

func randToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func Cors() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
        c.Next()
    }
}

func main() {
    router :=gin.Default()
    router.Use(Cors())

    /* view */
    router.StaticFile("/bootstrap.min.css","./bower_components/bootstrap/dist/css/bootstrap.min.css")
    router.StaticFile("/bootstrap.min.css.map","./bower_components/bootstrap/dist/css/bootstrap.min.css.map")
    router.StaticFile("/bootstrap-theme.min.css","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css")
    router.StaticFile("/bootstrap-theme.min.css.map","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css.map")
    router.LoadHTMLFiles("templates/index.html")
    // router.GET("/index", func(c *gin.Context) {
    //     c.HTML(http.StatusOK, "index.html", gin.H{
    //         "title": "Main website",
    //     })
    // })

    /* api */
    v1 := router.Group("api/v1")
    {
        v1.POST("/login",Login)
        v1.POST("/users/create", CreateUser)
    }

    router.Run(":8080")
}

type LoginForm struct {
    Account     string `form:"account" json:"account"`
    Passwd string `form:"passwd" json:"passwd"`
}


var loginToken map[string]string

func Login(c *gin.Context) {
    db := InitDb()
    defer db.Close()

    var loginData LoginForm
    c.Bind(&loginData)
    if loginData.Account != "" && loginData.Passwd != "" {
        var user Users
        db.Where("account = ? AND passwd = ?",loginData.Account , loginData.Passwd).First(&user)
        if user.Account != "" {
            loginToken[user.Account] = randToken()
            c.JSON(200, gin.H{"status":user})
        } else {
            c.JSON(404, gin.H{"status":"Not Found."})
        }
    }
    c.JSON(404, gin.H{"status":"Not Found."})
}

func CreateUser(c *gin.Context) {
    db := InitDb()
    defer db.Close()
    var user Users

    c.Bind(&user)
    c.JSON(200, gin.H{"account":user.Account,"passwd":user.Passwd})
    if user.Account != "" && user.Passwd != "" && user.Name != "" {
        db.Create(&user)
        c.JSON(201, gin.H{"success": user})
    } else {
        c.JSON(422, gin.H{"error": "Fields are empty"})
    }
}
