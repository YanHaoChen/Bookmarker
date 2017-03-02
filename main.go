package main

import (
    "time"
    "gopkg.in/gin-gonic/gin.v1"
    "net/http"
)

type Users struct {
    gorm.Model
    Account string `gorm:"primary_key" form:"account" json:"account"`
    Passwd  string `gorm:"not null" form:"passwd" json:"passwd"`
    Books []Books `gorm:"AssociationForeignKey:UserID" form:"books" json:"books"`
}
type Books struct {
    gorm.Model
    ID int `gorm:"AUTO_INCREMENT" form:"id" json:"id"`
    UserID int `form:"userID" json:"userID"`
    Name string `gorm:"not null" form:"name" json:"name"`
    Pages  int `gorm:"not null" form:"pages" json:"pages"`
    CreatedAt time.Time `form:"createdAt" json:"createdAt`
    Records []BookRecords `gorm:"AssociationForeignKey:BookID" form:"records" json:"records"`
}

type BookRecords struct {
    gorm.Model
    ID int `gorm:"AUTO_INCREMENT" form:"id" json:"id"`
    BookID int
    pages int `gorm:"not null" form:"pages" json:"pages"`
    CreatedAt time.Time `form:"createdAt" json:"createdAt`
}

func main() {
    router :=gin.Default()

    /* view */
    router.StaticFile("/bootstrap.min.css","./bower_components/bootstrap/dist/css/bootstrap.min.css")
    router.StaticFile("/bootstrap.min.css.map","./bower_components/bootstrap/dist/css/bootstrap.min.css.map")
    router.StaticFile("/bootstrap-theme.min.css","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css")
    router.StaticFile("/bootstrap-theme.min.css.map","./bower_components/bootstrap/dist/css/bootstrap-theme.min.css.map")
    router.LoadHTMLFiles("templates/index.html")
    router.GET("/index", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", gin.H{
            "title": "Main website",
        })
    })

    /* api */


    router.Run(":8080")
}
