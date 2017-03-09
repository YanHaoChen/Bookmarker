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
    Account string `gorm:";unique" form:"account" json:"account"`
    Passwd  string `gorm:"not null" form:"passwd" json:"passwd"`
    Name string `gorm:"not null" form:"name" json:"name"`
    Email string `gorm:"not null" form:"email" json:"email"`
    Books []Books `gorm:"ForeignKey:UserID" form:"books" json:"books"`
}
type Books struct {
    gorm.Model
    UserID uint `form:"userID" json:"userID"`
    Name string `gorm:"not null;unique" form:"name" json:"name"`
    Category string `gorm:"not null" form:"category" json:"category"`
    Pages  int `gorm:"not null" form:"pages" json:"pages"`
    Records []BookRecords `gorm:"ForeignKey:BookID" form:"records" json:"records"`
    Description string `gorm:"default:''" form:"description" json:"description"`
}
type BookRecords struct {
    gorm.Model
    BookID uint `gorm:"not null" form:"bookID" json:"bookID"`
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
    db := InitDb()
    defer db.Close()
    router = gin.Default()
    router.Use(Cors())
    loginToken = make(map[string]uint)

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
        /* user */
        /* account:string passwd:string */
        v1.POST("/login",Login)
        /* token:string */
        v1.POST("/logout", Logout)
        /* Account:string passwd:string name:string email:string */
        v1.POST("/users/create", CreateUser)
        /* token:string */
        v1.GET("/users/info", UserInfo)

        /* book */
        /* token:string name:string category:string pages:int description:string */
        v1.POST("/books/create", CreateBook)
        /* token:string */
        v1.GET("/books/infos", BookInfos)
        /* token:string bookID:uint name:string category:string pages:int description:string */
        v1.PUT("/books/update",UpdateBook)
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
    c.JSON(200, gin.H{"status":"Clear"})
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

func UserInfo(c *gin.Context) {
    var token Token
    c.Bind(&token)
    userID, exist := loginToken[token.Value]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    db :=InitDb()
    defer db.Close()
    var user Users
    if err := db.Table("users").Where("id = ?",userID).First(&user).Error; err != nil {
        c.JSON(422, gin.H{"error":"There are some things wrong."})
    } else {
        c.JSON(201, gin.H{"account":user.Account, "name":user.Name, "email":user.Email})
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

    if createBookValues.Name != "" && createBookValues.Category != "" && createBookValues.Pages > 0 {
        db := InitDb()
        defer db.Close()
        book := Books{
            Name : createBookValues.Name,
            Category : createBookValues.Category,
            Pages : createBookValues.Pages,
            Description : createBookValues.Description,
        }

        var user Users
        if err := db.Table("users").Where("id = ?",userID).First(&user).Error; err != nil {
            c.JSON(422, gin.H{"error":"Not found the user."})
        } else {
            if err := db.Model(&user).Association("Books").Append(&book).Error; err != nil {
                c.JSON(422, gin.H{"error":"Can't append this book."})
            } else {
                c.JSON(201, gin.H{"status":"Created"})
            }
        }
    } else {
        c.JSON(422, gin.H{"error":"There are some empty fields."})
    }
}

func BookInfos(c *gin.Context)  {
    var token Token
    c.Bind(&token)
    userID, exist := loginToken[token.Value]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    db :=InitDb()
    defer db.Close()
    var user Users
    if err := db.Table("users").Where("id = ?",userID).First(&user).Error; err != nil {
        c.JSON(422, gin.H{"error":"There are some things wrong."})
    } else {
        books :=[]Books{}
        db.Model(&user).Related(&books, "Books")
        for index , _ := range books {
            db.Model(&books[index]).Related(&books[index].Records, "Records")
        }
        c.JSON(201, gin.H{"books":books})

    }
}

type UpdateBookValues struct {
    Token string `form:"token" json:"token"`
    BookID uint `form:"bookID" json:"bookID"`
    Name string `form:"name" json:"name"`
    Category string `form:"category" json:"category"`
    Pages  int `form:"pages" json:"pages"`
    Description string `form:"description" json:"description"`
}

func UpdateBook(c *gin.Context)  {
    updateBookValues := UpdateBookValues{}
    c.Bind(&updateBookValues)

    userID, exist := loginToken[updateBookValues.Token]

    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if updateBookValues.Name != "" && updateBookValues.Category != "" && updateBookValues.Pages > 0 {
        db := InitDb()
        defer db.Close()
        book := Books{}
        db.Table("books").Where("id = ? and user_id = ?",updateBookValues.BookID, userID).First(&book)
        book.Name = updateBookValues.Name
        book.Category = updateBookValues.Category
        book.Pages = updateBookValues.Pages
        book.Description = updateBookValues.Description

        if err := db.Save(&book).Error; err != nil {
            c.JSON(422, gin.H{"error":"Can't update."})
        } else {
            c.JSON(201, gin.H{"status":"Updated"})
        }
    } else {
        c.JSON(422, gin.H{"error":"Can't find this book."})
    }

}
