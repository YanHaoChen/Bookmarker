package main

import (
    "crypto/rand"
	"fmt"
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
         c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
         c.Writer.Header().Set("Access-Control-Max-Age", "86400")
         c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
         c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
         c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
         c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

         if c.Request.Method == "OPTIONS" {
             fmt.Println("OPTIONS")
             c.AbortWithStatus(200)
         } else {
             c.Next()
         }
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

    /* api */
    v1 := router.Group("api/v1")
    {
        /* user */
        /* account:string passwd:string */
        v1.POST("/login",Login)
        /* token:string */
        v1.POST("/auth", Auth)
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
        v1.PUT("/books/update", UpdateBook)
        /* token:string bookID:uint */
        v1.DELETE("/books/delete",DeleteBook)

        /* BookRecord */
        /* token:string bookID:uint pages:int note:string */
        v1.POST("/bookrecords/create",CreateBookRecord)
        /* token:string */
        v1.GET("/bookrecords/infos",BookRecordInfos)
        /* token:string bookID:uint recordID:uint pages:int note:string */
        v1.PUT("/bookrecords/update",UpdateBookRecord)
        /* token:string bookID:uint recordID:uint */
        v1.DELETE("/bookrecords/delete",DeleteBookRecord)


    }
    router.Run(":8080")
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

func Auth(c *gin.Context)  {
    token := Token{}
    c.Bind(&token)
    _, exist := loginToken[token.Value]
    if exist == true {
        c.JSON(202, gin.H{"status":"Accept"})

    } else {
        c.JSON(403, gin.H{"error":"No permission."})
    }
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
        c.JSON(200, gin.H{"account":user.Account, "name":user.Name, "email":user.Email})
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
            c.JSON(406, gin.H{"error":"There are some things wrong."})
        } else {
            if user.Account != "" {
                user.Name = updateUserParams.Name
                user.Email = updateUserParams.Email
                if err := db.Save(&user).Error; err != nil {
                    c.JSON(406, gin.H{"error":"Can't update this user."})
                } else {
                    c.JSON(202, gin.H{"status":"Update"})
                }
            } else {
                c.JSON(406, gin.H{"error":"Can't find this user."})
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
            c.JSON(406, gin.H{"error":"There are some things wrong."})
        } else {
            if user.Account != "" {
                user.Passwd = upadateUserPasswdParams.Newpasswd
                if err := db.Save(&user).Error; err != nil {
                    c.JSON(406, gin.H{"error":"Can't update this user."})
                } else {
                    c.JSON(202, gin.H{"status":"Update"})
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
            c.JSON(406, gin.H{"error":"Not found the user."})
        } else {
            if err := db.Model(&user).Association("Books").Append(&book).Error; err != nil {
                c.JSON(406, gin.H{"error":"Can't append this book."})
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
        c.JSON(406, gin.H{"error":"There are some things wrong."})
    } else {
        books :=[]Books{}
        db.Model(&user).Related(&books, "Books")
        for index , _ := range books {
            db.Model(&books[index]).Related(&books[index].Records, "Records")
        }
        c.JSON(200, gin.H{"books":books})

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
    updateBookParams := UpdateBookParams{}
    c.Bind(&updateBookParams)

    userID, exist := loginToken[updateBookParams.Token]

    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if updateBookParams.Name != "" && updateBookParams.Category != "" && updateBookParams.Pages > 0 {
        db := InitDb()
        defer db.Close()
        book := Books{}
        db.Table("books").Where("id = ? and user_id = ?",updateBookParams.BookID, userID).First(&book)
        db.Model(&book).Related(&book.Records, "Records")
        read := 0

        for _, record := range book.Records {
            read += record.Pages
        }

        if updateBookParams.Pages < read {
            c.JSON(406, gin.H{"error":" Pages of This book has been read over the set of pages."})
            return
        }

        book.Name = updateBookParams.Name
        book.Category = updateBookParams.Category
        book.Pages = updateBookParams.Pages
        book.Description = updateBookParams.Description

        if err := db.Save(&book).Error; err != nil {
            c.JSON(406, gin.H{"error":"Can't update."})
        } else {
            c.JSON(202, gin.H{"status":book})
        }
    } else {
        c.JSON(422, gin.H{"error":"Can't find this book."})
    }

}
type DeleteBookParams struct {
    Token string `form:"token" json:"token"`
    BookID int `form:"bookID" json:"bookID"`
}

func DeleteBook(c *gin.Context)  {
    deleteBookParams := DeleteBookParams{}
    c.Bind(&deleteBookParams)
    userID, exist := loginToken[deleteBookParams.Token]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if deleteBookParams.BookID  > 0  {
        db := InitDb()
        defer db.Close()
        if err := db.Table("books").Where("user_id=? and id=?",userID , deleteBookParams.BookID).Delete(&Books{}).Error; err != nil {
            c.JSON(406, gin.H{"error":"Can't find this book."})
        } else {
            db.Table("book_records").Where("book_id=?",deleteBookParams.BookID).Delete(&Books{})
            c.JSON(202, gin.H{"status": "Delete"})
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
            c.JSON(406, gin.H{"error":"Can't find this book."})
        } else {

            db.Model(&book).Related(&book.Records, "Records")

            read := 0

            for _, record := range book.Records {
                read += record.Pages
            }
            if read + createBookRecordParams.Pages > book.Pages {
                c.JSON(406, gin.H{"error":"Over pages of this book."})
                return
            }
            bookRecord := BookRecords {
                Pages: createBookRecordParams.Pages,
                Note: createBookRecordParams.Note,
            }
            if err := db.Model(&book).Association("Records").Append(&bookRecord).Error; err != nil {
                c.JSON(406, gin.H{"error":"Can't append this BookRecord."})
            } else {
                c.JSON(201, gin.H{"status":"Created"})
            }
        }
    } else {
        c.JSON(422, gin.H{"error":"There are some empty fields."})
    }
}

func BookRecordInfos(c *gin.Context)  {
    token := Token{}
    c.Bind(&token)
    userID, exist := loginToken[token.Value]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    db := InitDb()
    defer db.Close()
    books := []Books{}
    if err := db.Table("books").Where("user_id = ?", userID).Find(&books).Error; err != nil {
        c.JSON(406, gin.H{"error":"There are some things wrong."})
    } else {
        records := []BookRecords{}
        for _, book := range books {
            db.Model(&book).Related(&book.Records, "Records")
            for _, record := range book.Records {
                records=append(records,record)
            }
        }
        c.JSON(200, gin.H{"bookrecords": records})
    }
}

type UpdateBookRecordParams struct {
    Token string `form:"token" json:"token"`
    BookID uint `form:"bookID" json:"bookID"`
    RecordID uint `form:"recordID" json:"recordID"`
    Pages int `form:"pages" json:"pages"`
    Note string `form:"note" json:"note"`
}

func UpdateBookRecord(c *gin.Context)  {
    updateBookRecordParams := UpdateBookRecordParams{}
    c.Bind(&updateBookRecordParams)
    userID, exist := loginToken[updateBookRecordParams.Token]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if updateBookRecordParams.BookID > 0 && updateBookRecordParams.RecordID > 0 && updateBookRecordParams.Pages >= 0 {
        db := InitDb()
        defer db.Close()
        book := Books{}
        if err := db.Table("books").Where("user_id = ? and id = ?", userID, updateBookRecordParams.BookID).First(&book).Error; err != nil {
            c.JSON(405, gin.H{"error":"Can't find this book."})
        } else {
            updateRecord := BookRecords{}
            if err := db.Table("book_records").Where("id = ?",updateBookRecordParams.RecordID).First(&updateRecord).Error; err != nil {
                c.JSON(406, gin.H{"error":"Can't find this record."})
            }else {
                db.Model(&book).Related(&book.Records,"Records")
                read := 0
                for _,record := range book.Records {
                    read += record.Pages
                }
                read -= updateRecord.Pages
                if read + updateBookRecordParams.Pages > book.Pages {
                    c.JSON(406, gin.H{"error":"Over pages of this book."})
                } else {
                    updateRecord.Pages=updateBookRecordParams.Pages
                    updateRecord.Note=updateBookRecordParams.Note
                    if err := db.Table("book_records").Save(&updateRecord).Error; err != nil {
                        c.JSON(406, gin.H{"error":"Can't update this record."})
                    }else {
                        c.JSON(202,gin.H{"status":"Updated"})
                    }
                }
            }
        }
    } else {
        c.JSON(422, gin.H{"error": "There are some empty fields."})
    }
}

type DeleteBookRecordParams struct {
    Token string `form:"token" json:"token"`
    BookID uint `form:"bookID" json:"bookID"`
    RecordID uint `form:"recordID" json:"recordID"`
}

func DeleteBookRecord(c *gin.Context)  {
    deleteBookRecordParams := DeleteBookRecordParams{}
    c.Bind(&deleteBookRecordParams)
    userID, exist := loginToken[deleteBookRecordParams.Token]
    if exist == false {
        c.JSON(403, gin.H{"error":"No permission."})
        return
    }
    if deleteBookRecordParams.BookID >0 && deleteBookRecordParams.RecordID > 0 {
        db := InitDb()
        defer db.Close()
        book := Books{}
        if err := db.Table("books").Where("user_id=? and id=?",userID, deleteBookRecordParams.BookID).First(&book).Error; err != nil {
            c.JSON(406, gin.H{"error":"Can't find this book."})
        } else {
            record := BookRecords{}
            if err := db.Table("book_records").Where("book_id=? and id=?",deleteBookRecordParams.BookID,deleteBookRecordParams.RecordID).First(&record).Error; err != nil {
                c.JSON(406, gin.H{"error":"Can't find this record."})
            } else {
                if err := db.Delete(&record).Error; err != nil {
                    c.JSON(406, gin.H{"error":"Can't delete this record."})
                } else {
                    c.JSON(202, gin.H{"status":"Deleted"})
                }
            }
        }
    } else {
        c.JSON(422, gin.H{"error":"There are some empty fields."})
    }
}
