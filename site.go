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

type ContactInfo struct {
	ID          string `sql:"id" json:"id"`
	FirstName   string `sql:"first_name" json:"first_name"`
	LastName    string `sql:"last_name" json:"last_name"`
	Phone       int    `sql:"phone" json:"phone"`
	OfficePhone int    `sql:"office_phone" json:"office_phone"`
	City        string `sql:"city" json:"city"`
	State       string `sql:"state" json:"state"`
	Zip         string `sql:"zip" json:"zip"`
	Enabled     bool   `sql:"enabled" json:"enabled"`
}

func NewContact() ContactInfo {
	contact := ContactInfo{}
	return contact
}

func (ac *appContext) ShowIndex(c *gin.Context) {
	ac.Log.Msg(1, "in context show index")

	query :=
		` select 
			id, first_name , last_name, phone, office_phone, city, state, zip, enabled 
		from contacts 
		where 
			enabled
			order by 3,2`

	rows, err := ac.DB.Query(query)
	if check := ac.DBErrorCheck(err, query, c); check == false {
		ac.Log.Msg(1, "db error")
		return
	}
	defer rows.Close()

	var contacts []ContactInfo

	ac.Log.Msg(1, "before assignment")
	for rows.Next() {
		contact := NewContact()

		err := rows.Scan(&contact.ID, &contact.FirstName, &contact.LastName, &contact.Phone, &contact.OfficePhone,
			&contact.City, &contact.State, &contact.Zip, &contact.Enabled)
		if err != nil {
			ac.Log.Msg(3, fmt.Sprintf("Error scanning row: %s", err.Error()))
		}

		ac.Log.Msg(1, fmt.Sprintf("%#v", contact))

		contacts = append(contacts, contact)
	}

	c.HTML(http.StatusOK, "main/index", gin.H{
		"contacts": contacts,
	})

}

func (ac *appContext) uploadContact(c *gin.Context) {
	var form formPostData

	if err := c.Bind(&form); err != nil {
		fmt.Println("bind error" + err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": "ok"})
	fmt.Printf("%+v", form)
	query := "insert into contacts (first_name, last_name, phone, office_phone, " +
		"city, state, zip) " +
		"values ($1, $2, $3, $4, $5, $6, $7)"
	_, err := ac.DB.Exec(query, &form.FirstName, &form.LastName, &form.Phone,
		&form.OfficePhone, &form.City, &form.State, &form.Zip)
	ac.DBErrorCheck(err, query, c)

}
func (ac *appContext) editContact(c *gin.Context) {

}
func (ac *appContext) saveContact(c *gin.Context) {

}
func (ac *appContext) deleteContact(c *gin.Context) {

}
/*func (ac *appContext) handlerPostData(c *gin.Context) {
	ac.Log.Msg(1, "in context handler post data")
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
