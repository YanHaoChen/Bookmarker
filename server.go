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
    Books []Books `gorm:"ForeignKey:ID;AssociationForeignKey:UserID" form:"books" json:"books"`
}
type Books struct {
    gorm.Model
    UserID uint `form:"userID" json:"userID"`
    Name string `gorm:"not null" form:"name" json:"name"`
    Category string `gorm:"not null" form:"category" json:"category"`
    Pages  int `gorm:"not null" form:"pages" json:"pages"`
    Records []BookRecords `gorm:"ForeignKey:ID;AssociationForeignKey:BookID" form:"records" json:"records"`
    Description string `gorm:"default:''" form:"description" json:"description"`
}
type BookRecords struct {
    gorm.Model
    BookID uint
    Pages int `gorm:"not null" form:"pages" json:"pages"`
    Note string `gorm:"default:''" form:"note" json:"note"`
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
var loginToken map[string]uint

func main() {
    router = gin.Default()
    router.Use(Cors())
    loginToken = make(map[string]uint)

    /* bootstrap */
    router.StaticFile("/bootstrap/css/bootstrap.min.css","./bower_components/bootstrap/dist/css/bootstrap.min.css")
    router.StaticFile("/bootstrap/css/bootstrap.min.css.map","./bower_components/bootstrap/dist/css/bootstrap.min.css.map")
    router.StaticFile("/bootstrap/css/bootstrap-theme.min.css","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css")
    router.StaticFile("/bootstrap/css/bootstrap-theme.min.css.map","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css.map")
    router.StaticFile("/bootstrap/css/signin.css","./bower_components/bootstrap/dist/css/signin.css")
    /* js */
    router.StaticFile("/js/vue/vue.min.js","./bower_components/vue/dist/vue.min.js")
    router.StaticFile("/js/vue-resource/vue-resource.min.js","./bower_components/vue-resource/dist/vue-resource.min.js")

    /* view */
    view := router.Group("view")
    {
        view.GET("/welcome", Welcome)
        view.GET("/dashboard", Dashboard)
    }
    /* api */
    v1 := router.Group("api/v1")
    {
        /* user */
        v1.POST("/login",Login)
        v1.POST("/logout", Logout)
        v1.POST("/users/create", CreateUser)
        // delete

        /* book */
        v1.POST("books/create", CreateBook)
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

type LoginForm struct {
    Account     string `form:"account" json:"account"`
    Passwd string `form:"passwd" json:"passwd"`
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
            loginToken[token] = user.ID
            c.JSON(200, gin.H{"token":token})
        } else {
            c.JSON(404, gin.H{"status":"Not Found."})
        }
    }
}

type Token struct {
     Value string `form:"token" json:"token"`
}

func Logout(c *gin.Context) {
    var token Token
    c.Bind(&token)
    delete(loginToken, token.Value)
    c.JSON(200, gin.H{"status":"clear"})
}

func CreateUser(c *gin.Context) {
    db := InitDb()
    defer db.Close()
    var user Users

    c.Bind(&user)
    if user.Account != "" && user.Passwd != "" && user.Name != "" && user.Email != "" {
        if err := db.Create(&user).Error; err != nil {
            c.JSON(422, gin.H{"error":"There are some things wrong."})
        } else {
            c.JSON(201, gin.H{"success": user})
        }
    } else {
        c.JSON(422, gin.H{"error": "There are some empty fields."})
    }
}

type CreateBookValues struct {
     Token string `form:"token" json:"token"`
     Name string `form:"name" json:"name"`
     Category string `form:"category" json:"category"`
     Pages  int `form:"pages" json:"pages"`
     Description string `form:"description" json:"description"`
}

func CreateBook(c *gin.Context)  {
    var createBookValues CreateBookValues
    c.Bind(&createBookValues)

    userID, exist := loginToken[createBookValues.Token]

    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }

    if createBookValues.Name != "" && createBookValues.Category != "" && createBookValues.Category != "" {
        db := InitDb()
        defer db.Close()
        var book Books
        book.Name = createBookValues.Name
        book.UserID = userID
        book.Category = createBookValues.Category
        book.Pages = createBookValues.Pages
        book.Description = createBookValues.Description
        if err := db.Create(&book).Error; err != nil {
            c.JSON(422, gin.H{"error":"There are some things wrong."})
        } else {
            c.JSON(201, gin.H{"status":"created"})
        }
    } else {
        c.JSON(422, gin.H{"error":"There are some empty fields."})
    }
}
