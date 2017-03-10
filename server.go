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
    Name string `gorm:"not null" form:"name" json:"name"`
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
        /* token:string name:string email:string */
        v1.PUT("/users/update",UpdateUser)
        /* token:string expasswd:string newpasswd:string */
        v1.PUT("/users/updatepasswd",UpdateUserPasswd)


        /* book */
        /* token:string name:string category:string pages:int description:string */
        v1.POST("/books/create", CreateBook)
        /* token:string */
        v1.GET("/books/infos", BookInfos)
        /* token:string bookID:uint name:string category:string pages:int description:string */
        v1.PUT("/books/update",UpdateBook)
        /* token:string bookID:uint */
        v1.DELETE("/books/delete",DeleteBook)

        /* BookRecord */
        /* token:string bookID:string pages:int note:string */
        v1.POST("/bookrecords/create",CreateBookRecord)


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
    token := Token{}
    c.Bind(&token)
    delete(loginToken, token.Value)
    c.JSON(200, gin.H{"status":"Clear"})
}

func CreateUser(c *gin.Context) {
    db := InitDb()
    defer db.Close()
    user := Users{}

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
    token := Token{}
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

type UpdateUserParams struct {
    Token string `form:"token" json:"token"`
    Name string `form:"name" json:"name"`
    Email string `form:"email" json:"email"`
}

func UpdateUser(c *gin.Context) {
    updateUserParams := UpdateUserParams{}
    c.Bind(&updateUserParams)
    userID, exist := loginToken[updateUserParams.Token]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }

    if updateUserParams.Name != "" && updateUserParams.Email != "" {
        db :=InitDb()
        defer db.Close()
        user := Users{}
        if err := db.Table("users").Where("id = ?",userID).First(&user).Error; err != nil {
            c.JSON(422, gin.H{"error":"There are some things wrong."})
        } else {
            if user.Account != "" {
                user.Name = updateUserParams.Name
                user.Email = updateUserParams.Email
                if err := db.Save(&user).Error; err != nil {
                    c.JSON(422, gin.H{"error":"Can't update this user."})
                } else {
                    c.JSON(201, gin.H{"status":"Update"})
                }
            } else {
                c.JSON(422, gin.H{"error":"Can't find this user."})
            }
        }
    } else {
        c.JSON(422, gin.H{"error":"There are some empty fields."})
    }
}

type UpdateUserPasswdParams struct {
    Token string `form:"token" json:"token"`
    Expasswd string `form:"expasswd" json:"expasswd"`
    Newpasswd string `form:"newpasswd" json:"newpasswd"`
}

func UpdateUserPasswd(c *gin.Context)  {
    upadateUserPasswdParams := UpdateUserPasswdParams{}
    c.Bind(&upadateUserPasswdParams)

    userID, exist := loginToken[upadateUserPasswdParams.Token]

    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if upadateUserPasswdParams.Expasswd != "" && upadateUserPasswdParams.Newpasswd != "" {
        db :=InitDb()
        defer db.Close()
        user := Users{}
        if err := db.Table("users").Where("id = ? and passwd = ?",userID, upadateUserPasswdParams.Expasswd).First(&user).Error; err != nil {
            c.JSON(422, gin.H{"error":"There are some things wrong."})
        } else {
            if user.Account != "" {
                user.Passwd = upadateUserPasswdParams.Newpasswd
                if err := db.Save(&user).Error; err != nil {
                    c.JSON(422, gin.H{"error":"Can't update this user."})
                } else {
                    c.JSON(201, gin.H{"status":"Update"})
                }
            } else {
                c.JSON(422, gin.H{"error":"Can't find this user."})
            }
        }
    } else {
        c.JSON(422, gin.H{"error":"There are some empty fields."})
    }

}

type CreateBookParams struct {
     Token string `form:"token" json:"token"`
     Name string `form:"name" json:"name"`
     Category string `form:"category" json:"category"`
     Pages  int `form:"pages" json:"pages"`
     Description string `form:"description" json:"description"`
}

func CreateBook(c *gin.Context)  {
    var CreateBookParams CreateBookParams
    c.Bind(&CreateBookParams)

    userID, exist := loginToken[CreateBookParams.Token]

    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }

    if CreateBookParams.Name != "" && CreateBookParams.Category != "" && CreateBookParams.Pages > 0 {
        db := InitDb()
        defer db.Close()
        book := Books{
            Name : CreateBookParams.Name,
            Category : CreateBookParams.Category,
            Pages : CreateBookParams.Pages,
            Description : CreateBookParams.Description,
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

type UpdateBookParams struct {
    Token string `form:"token" json:"token"`
    BookID uint `form:"bookID" json:"bookID"`
    Name string `form:"name" json:"name"`
    Category string `form:"category" json:"category"`
    Pages  int `form:"pages" json:"pages"`
    Description string `form:"description" json:"description"`
}

func UpdateBook(c *gin.Context)  {
    UpdateBookParams := UpdateBookParams{}
    c.Bind(&UpdateBookParams)

    userID, exist := loginToken[UpdateBookParams.Token]

    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if UpdateBookParams.Name != "" && UpdateBookParams.Category != "" && UpdateBookParams.Pages > 0 {
        db := InitDb()
        defer db.Close()
        book := Books{}
        db.Table("books").Where("id = ? and user_id = ?",UpdateBookParams.BookID, userID).First(&book)
        book.Name = UpdateBookParams.Name
        book.Category = UpdateBookParams.Category
        book.Pages = UpdateBookParams.Pages
        book.Description = UpdateBookParams.Description

        if err := db.Save(&book).Error; err != nil {
            c.JSON(422, gin.H{"error":"Can't update."})
        } else {
            c.JSON(201, gin.H{"status":"Updated"})
        }
    } else {
        c.JSON(422, gin.H{"error":"Can't find this book."})
    }

}
type DeleteBookParams struct {
    Token string `form:"token" json:"token"`
    BookID string `form:"bookID" json:"bookID"`
}

func DeleteBook(c *gin.Context)  {
    deleteBookParams := DeleteBookParams{}
    c.Bind(&deleteBookParams)
    userID, exist := loginToken[deleteBookParams.Token]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if deleteBookParams.BookID != "" {
        db := InitDb()
        defer db.Close()
        if err := db.Table("books").Where("user_id=? and id=?",userID , deleteBookParams.BookID).Delete(&Books{}).Error; err != nil {
            c.JSON(422, gin.H{"error":"Can't find this book."})
        } else {
            db.Table("book_records").Where("book_id=?",deleteBookParams.BookID).Delete(&Books{})
            c.JSON(201, gin.H{"status": "Delete"})
        }
    } else {
        c.JSON(422, gin.H{"error": "There are some empty fields."})
    }
}

type CreateBookRecordParams struct {
    Token string `form:"token" json:"token"`
    BookID uint `form:"bookID" json:"bookID"`
    Pages int `form:"pages" json:"pages"`
    Note string `form:"note" json:"note"`
}

func CreateBookRecord(c *gin.Context) {
    createBookRecordParams := CreateBookRecordParams{}
    c.Bind(&createBookRecordParams)
    userID, exist := loginToken[createBookRecordParams.Token]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }

    if createBookRecordParams.BookID > 0 && createBookRecordParams.Pages >= 0 {
        db := InitDb()
        defer db.Close()
        book := Books{}
        if err := db.Table("books").Where("user_id=? and id =?",userID, createBookRecordParams.BookID).First(&book).Error; err != nil {
            c.JSON(422, gin.H{"error":"Can't find this book."})
        } else {

            db.Model(&book).Related(&book.Records, "Records")

            read := 0

            for _, record := range book.Records {
                read += record.Pages
            }
            if read + createBookRecordParams.Pages > book.Pages {
                c.JSON(422, gin.H{"error":"Over pages of this book."})
                return
            }
            bookRecord := BookRecords {
                Pages: createBookRecordParams.Pages,
                Note: createBookRecordParams.Note,
            }
            if err := db.Model(&book).Association("Records").Append(&bookRecord).Error; err != nil {
                c.JSON(422, gin.H{"error":"Can't append this BookRecord."})
            } else {
                c.JSON(201, gin.H{"status":"Created"})
            }
        }
    } else {
        c.JSON(422, gin.H{"error":"There are some empty fields."})
    }
}
