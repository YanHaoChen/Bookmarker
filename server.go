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
    Email string `gorm:"not null" form:"email" json:"email"`
    Books []Books `gorm:"AssociationForeignKey:UserID" form:"books" json:"books"`
}
type Books struct {
    gorm.Model
    UserID int `form:"userID" json:"userID"`
    Name string `gorm:"not null" form:"name" json:"name"`
    Category string `gorm:"not null" form:"category" json:"category"`
    Pages  int `gorm:"not null" form:"pages" json:"pages"`
    Records []BookRecords `gorm:"AssociationForeignKey:BookID" form:"records" json:"records"`
    Description string `form:"description" json:"description"`
}
type BookRecords struct {
    gorm.Model
    BookID int
    Pages int `gorm:"not null" form:"pages" json:"pages"`
    Note string `form:"note" json:"note"`
}
type LoginForm struct {
    Account     string `form:"account" json:"account"`
    Passwd string `form:"passwd" json:"passwd"`
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
	b := make([]byte, 15)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func Cors() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
        c.Next()
    }
}

var router *gin.Engine
var loginToken map[string]string

func main() {
    router = gin.Default()
    router.Use(Cors())
    loginToken = make(map[string]string)

    /* bootstrap */
    router.StaticFile("/bootstrap/css/bootstrap.min.css","./bower_components/bootstrap/dist/css/bootstrap.min.css")
    router.StaticFile("/bootstrap/css/bootstrap.min.css.map","./bower_components/bootstrap/dist/css/bootstrap.min.css.map")
    router.StaticFile("/bootstrap/css/bootstrap-theme.min.css","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css")
    router.StaticFile("/bootstrap/css/bootstrap-theme.min.css.map","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css.map")
    router.StaticFile("/bootstrap/css/signin.css","./bower_components/bootstrap/dist/css/signin.css")
    router.StaticFile("/bootstrap/css/dashboard.css","./bower_components/bootstrap/dist/css/dashboard.css")

    /* js */
    router.StaticFile("/js/vue/vue.min.js","./bower_components/vue/dist/vue.min.js")
    router.StaticFile("/js/vue-resource/vue-resource.min.js","./bower_components/vue-resource/dist/vue-resource.min.js")
    router.StaticFile("/js/jsSHA/sha.js","./bower_components/jsSHA/src/sha.js")
    /* view */
    view := router.Group("view")
    {
        view.GET("/welcome", Welcome)
        view.GET("/dashboard", Dashboard)
    }
    /* api */
    v1 := router.Group("api/v1")
    {
        v1.POST("/login",Login)
        v1.POST("/users/create", CreateUser)
    }
    router.Run(":8080")
}

func Welcome(c *gin.Context) {
    router.LoadHTMLFiles("templates/welcome.html")
    c.HTML(http.StatusOK, "welcome.html", gin.H{
    })
}

func Dashboard(c *gin.Context) {
    router.LoadHTMLFiles("templates/dashboard.html")
    c.HTML(http.StatusOK, "dashboard.html", gin.H{
    })
}


func Login(c *gin.Context) {
    db := InitDb()
    defer db.Close()

    var loginData LoginForm
    c.Bind(&loginData)
    if loginData.Account != "" && loginData.Passwd != "" {
        var user Users
        db.Where("account = ? AND passwd = ?",loginData.Account , loginData.Passwd).First(&user)
        if user.Account != "" {
            token := randToken()
            loginToken[token] = user.Account
            c.JSON(200, gin.H{"token":token})
        } else {
            c.JSON(404, gin.H{"status":"Not Found."})
        }
    }
}

func CreateUser(c *gin.Context) {
    db := InitDb()
    defer db.Close()
    var user Users

    c.Bind(&user)
    if user.Account != "" && user.Passwd != "" && user.Name != "" && user.Email != "" {
        db.Create(&user)
        c.JSON(201, gin.H{"success": user})
    } else {
        c.JSON(422, gin.H{"error": "There are some empty fields ."})
    }
}
