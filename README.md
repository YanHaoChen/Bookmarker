# Bookmarker

紀錄閱讀進度的專案，並透過網頁的方式呈現。此專案主要是學習取向，學習前、後端分離的 Web service。

* 前端：Vue、Bootstrap（無使用到 jQuery 的部分）

* 後端：gin（Back-end Server）、gorm（ORM）
* 版本控制：進一步熟悉 git 的分工操作

前端 Vue 的部分，還沒有很完善的架構，只確定運作正常，待往後繼續調整。

## Front-end Server

直接將資料夾 [front-end](https://github.com/YanHaoChen/Bookmarker/tree/master/front-end) 放入 Apache Server 上即可。

>  

## Back-end Server

### 程式
[server.go](https://github.com/YanHaoChen/Bookmarker/tree/master/back-end)


### 需下載的模組

* github.com/gin-gonic/gin
* github.com/jinzhu/gorm
* github.com/mattn/go-sqlite3

### 執行
```shell
go run server.go
```

## API

### User
Name | Method | URL | Params
--- | --- | --- | ---
Login | POST | /login | account:string passwd:string
Auth | POST | /auth |token:string
Logout | POST | /logout | token:string
CreateUser | POST | /users/create |Account:string passwd:string name:string email:string
UserInfo | GET | /users/info?token= | token:string
UpdateUser | PUT | /users/update | token:string name:string email:string
UpdateUserPasswd | PUT | /users/updatepasswd | token:string expasswd:string newpasswd:string

### Book
Name | Method | URL | Params
--- | --- | --- | ---
CreateBook | POST | /books/create |token:string name:string category:string pages:int description:string
BookInfos | GET | /books/infos?token= | token:string
UpdateBook | PUT | /books/update |token:string bookID:uint name:string category:string pages:int description:string
DeleteBook | DELETE | /books/delete | token:string bookID:uint

### BookRecord
Name | Method | URL | Params
--- | --- | --- | ---
CreateBookRecord | POST | /bookrecords/create |token:string bookID:uint pages:int note:string
BookRecordInfos | GET | /bookrecords/infos?token= | token:string
UpdateBookRecord | PUT | /bookrecords/update |token:string bookID:uint recordID:uint pages:int note:string
DeleteBookRecord | DELETE | /bookrecords/delete | token:string bookID:uint recordID:uint


## Overview
![Overview](https://github.com/YanHaoChen/Bookmarker/blob/master/img/overview.png?raw=true)

## Books
![Overview](https://github.com/YanHaoChen/Bookmarker/blob/master/img/books.png?raw=true)

## 目前進度：80%

項目 | 進度
--- | ---
Overview | 100%
Books | 100%
Setting | 0%



