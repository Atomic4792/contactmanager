package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type formPostData struct {
	FirstName   string `form:"firstName"`
	LastName    string `form:"lastName"`
	Phone       int    `form:"phone"`
	OfficePhone int    `form:"officePhone"`
	City        string `form:"city"`
	State       string `form:"state"`
	Zip         string `form:"zip"`
}

func (ac *appContext) ShowIndex(c *gin.Context) {
	ac.Log.LogMsg(1, "in context show index")
	c.HTML(http.StatusOK, "main/index", gin.H{"contactsInfo": ac.contactsInfo})

}
func (ac *appContext) uploadContact(c *gin.Context) {
	var form formPostData

	if err := c.Bind(&form); err != nil {
		fmt.Println("bind error" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": "ok"})
	fmt.Printf("%+v",form)
	query := "insert into contacts (first_name, last_name, phone, office_phone, " +
		"city, state, zip) " +
		"values ($1, $2, $3, $4, $5, $6, $7)"
	_, err := ac.DB.Exec(query, &form.FirstName, &form.LastName, &form.Phone,
		&form.OfficePhone, &form.City, &form.State, &form.Zip)
	ac.DBErrorCheck(err, query)



}


/*func (ac *appContext) handlerPostData(c *gin.Context) {
	ac.Log.LogMsg(1, "in context handler post data")
	var form formPostData
	if err := c.Bind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	query := "insert into contacts (first_name, last_name, phone, office_phone, " +
		"city, state, zip) " +
		"values ($1, $2, $3, $4, $5, $6, $7)"
	_, err := ac.DB.Exec(query, &form.FirstName, &form.LastName, &form.Phone,
		&form.OfficePhone, &form.City, &form.State, &form.Zip)
	ac.DBErrorCheck(err, query)
	ac.LoadAppDefaults()
		c.String(http.StatusOK,fmt.Sprintln("all good bro"))
}

*/
